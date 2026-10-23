#!/usr/bin/env python3
# Copyright 2026 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import re
from dataclasses import dataclass

NAMESPACE = "cm-kernel-ci"
API = f"/apis/chaos-mesh.org/v1alpha1/namespaces/{NAMESPACE}/kernelchaos"
ROLES = ("target", "control", "bpfki", "daemon", "controller")
CONTAINER_NAMES = {"target": "worker", "control": "worker", "bpfki": "bpfki",
                   "daemon": "chaos-daemon", "controller": "chaos-mesh"}
IMAGE_ROLES = {"target": "worker", "control": "worker", "bpfki": "kernel",
               "daemon": "daemon", "controller": "controller"}
FINALIZER = "chaos-mesh/records"


@dataclass(frozen=True)
class Budget:
    active: int = 300
    expiration: int = 150
    delete: int = 90
    first_injection: int = 90
    duration: int = 120
    operation: int = 3
    operations: int = 8
    faults: int = 4

BUDGET = Budget()

def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def pod_identity(pod, role):
    metadata = pod["metadata"]
    require(not metadata.get("deletionTimestamp"), f"{role} Pod is deleting")
    require(pod["status"].get("phase") == "Running", f"{role} Pod is not Running")
    require(pod["status"].get("qosClass") == "Burstable", f"{role} Pod is not Burstable")
    name = CONTAINER_NAMES[role]
    matches = [item for item in pod["status"].get("containerStatuses", []) if item["name"] == name]
    require(len(matches) == 1, f"{role} has no unique container status")
    status = matches[0]
    require(status.get("ready") and "running" in status.get("state", {}), f"{role} is not ready")
    require(status.get("restartCount") == 0, f"{role} restarted")
    runtime_id = status["containerID"]
    require(re.fullmatch(r"containerd://[0-9a-f]{64}", runtime_id), "unexpected runtime container ID")
    spec = [item for item in pod["spec"]["containers"] if item["name"] == name]
    require(len(spec) == 1 and spec[0].get("imagePullPolicy") == "Never", "unexpected container image policy")
    return {"role": role, "namespace": metadata["namespace"], "pod": metadata["name"],
            "pod_uid": metadata["uid"], "container": name, "runtime_id": runtime_id,
            "pod_image_id": status["imageID"], "image": spec[0]["image"],
            "restart_count": status["restartCount"]}


def select_pods(items):
    selected = {}
    for role in ROLES:
        if role in ("target", "control"):
            matches = [pod for pod in items if pod["metadata"]["namespace"] == NAMESPACE
                       and pod["metadata"]["name"] == role]
        else:
            component = "controller-manager" if role == "controller" else "chaos-daemon"
            matches = [pod for pod in items if pod["metadata"]["namespace"] == "chaos-mesh"
                       and pod["metadata"].get("labels", {}).get("app.kubernetes.io/component") == component
                       and pod["metadata"].get("labels", {}).get("app.kubernetes.io/instance") == "chaos-mesh"]
        require(len(matches) == 1, f"no unique {role} Pod")
        selected[role] = matches[0]
    return selected


def validate_cri_identity(identity, status):
    container_id = identity["runtime_id"].removeprefix("containerd://")
    require(status["id"] == container_id and status["state"] == "CONTAINER_RUNNING", "CRI container mismatch")
    require(status["metadata"]["name"] == identity["container"], "CRI container name mismatch")
    labels = status["labels"]
    for key, value in (("io.kubernetes.pod.uid", identity["pod_uid"]),
                       ("io.kubernetes.pod.name", identity["pod"]),
                       ("io.kubernetes.pod.namespace", identity["namespace"]),
                       ("io.kubernetes.container.name", identity["container"])):
        require(labels.get(key) == value, f"CRI label mismatch {key}")


def image_binding(identity, inspect, image, expected):
    status, info = inspect["status"], inspect["info"]
    image_status, image_info = image["status"], image["info"]
    validate_cri_identity(identity, status)
    expected_pod_id = status.get("imageId") or status["imageRef"]
    require(identity["pod_image_id"] == expected_pod_id, "Pod imageID does not follow actual CRI ImageId/ImageRef")
    require(image_status["id"] == expected["config_digest"], "CRI image config differs from sealed input")
    allowed_refs = {image_status["id"], *image_status.get("repoDigests", [])}
    require(status["imageRef"] in allowed_refs, "CRI ImageRef is not owned by inspected image")
    require(expected["tag"] in image_status.get("repoTags", []), "sealed tag absent from CRI image")
    require(identity["image"] == expected["tag"], "Pod image tag differs from sealed input")
    require(image_info["imageSpec"]["rootfs"]["diff_ids"] == expected["diff_ids"], "CRI rootfs differs from sealed input")
    require(info["runtimeType"] == "io.containerd.runc.v2" and info["snapshotter"] == "overlayfs",
            "unexpected CRI runtime or snapshotter")
    return {"cri_image_id": image_status["id"], "cri_image_ref": status["imageRef"],
            "cri_container_image_id": status.get("imageId", ""),
            "repo_tags": image_status.get("repoTags", []), "repo_digests": image_status.get("repoDigests", []),
            "config_digest": expected["config_digest"], "diff_ids": expected["diff_ids"],
            "chain_id": image_info["chainID"], "task_pid": info["pid"]}


def spec(kind):
    require(kind in ("expiration", "delete"), "invalid fixed case")
    result = {"mode": "one", "selector": {"pods": {NAMESPACE: ["target"]}},
              "containerNames": ["worker"], "failKernRequest": {
                  "failtype": 0, "probability": 100, "times": 0, "callchain": [{"predicate": "true"}]}}
    if kind == "expiration":
        result["duration"] = str(BUDGET.duration) + "s"
    return result


def validate_object(obj, kind, uid):
    require(obj["metadata"]["uid"] == uid and obj["metadata"]["name"] == kind
            and obj["metadata"]["namespace"] == NAMESPACE, "experiment identity changed")
    annotations = obj["metadata"].get("annotations", {})
    require(not any("clean" in key or "pause" in key for key in annotations), "experiment has cleanup/pause annotation")
    actual = obj["spec"]
    expected = spec(kind)
    require(actual["mode"] == "one" and actual["selector"].get("pods") == expected["selector"]["pods"]
            and actual["containerNames"] == ["worker"], "experiment target changed")
    require(actual["selector"].get("namespaces") == [NAMESPACE], "defaulted target namespace differs")
    require(not any(value for key, value in actual["selector"].items() if key not in {"pods", "namespaces"}),
            "unexpected selector fields")
    request = actual["failKernRequest"]
    require(request["failtype"] == 0 and request["probability"] == 100 and request.get("times", 0) == 0
            and request["callchain"] == [{"predicate": "true"}] and not request.get("headers"), "fault specification changed")
    require(actual.get("duration") == expected.get("duration") and not actual.get("remoteCluster"),
            "experiment duration or cluster changed")


def record(obj):
    entries = obj.get("status", {}).get("experiment", {}).get("containerRecords", [])
    if not entries:
        return None
    require(len(entries) == 1, "experiment selected more than one record")
    value = entries[0]
    require(value["id"] == "cm-kernel-ci/target/worker" and value["selectorKey"] == ".", "wrong selected record")
    require(not any(event["type"] == "Failed" for event in value.get("events", [])), "controller recorded failed lifecycle action")
    require(value["injectedCount"] <= 1 and value["recoveredCount"] <= 1, "controller duplicated lifecycle action")
    return value


def event_succeeded(value, operation):
    return any(event.get("type") == "Succeeded" and event.get("operation") == operation
               for event in value.get("events", []))


def active(obj):
    value = record(obj)
    conditions = {item["type"]: item["status"] for item in obj.get("status", {}).get("conditions", [])}
    return (value is not None and obj["status"]["experiment"].get("desiredPhase") == "Run"
            and value["phase"] == "Injected" and value["injectedCount"] == 1 and value["recoveredCount"] == 0
            and event_succeeded(value, "Apply") and FINALIZER in obj["metadata"].get("finalizers", [])
            and conditions.get("Selected") == "True" and conditions.get("AllInjected") == "True"
            and conditions.get("AllRecovered") == "False" and not obj["metadata"].get("deletionTimestamp"))


def recovered(obj):
    value = record(obj)
    return (value is not None and obj["status"]["experiment"].get("desiredPhase") == "Stop"
            and value["phase"] == "Not Injected" and value["injectedCount"] == 1 and value["recoveredCount"] == 1
            and event_succeeded(value, "Apply") and event_succeeded(value, "Recover"))


def expiration_recovered(obj):
    conditions = {item["type"]: item["status"] for item in obj.get("status", {}).get("conditions", [])}
    return (recovered(obj) and conditions.get("AllRecovered") == "True" and conditions.get("AllInjected") == "False"
            and FINALIZER in obj["metadata"].get("finalizers", []) and not obj["metadata"].get("deletionTimestamp"))


