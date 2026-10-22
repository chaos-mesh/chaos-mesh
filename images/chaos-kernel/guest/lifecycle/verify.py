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

"""The two fixed KernelChaos lifecycle cases inside the sealed offline guest."""

import datetime
import errno
import hashlib
import json
import math
import os
from pathlib import Path
import re
import select
import selectors
import signal
import subprocess
import time
from urllib.parse import urlencode


NAMESPACE = "cm-kernel-ci"
API = f"/apis/chaos-mesh.org/v1alpha1/namespaces/{NAMESPACE}/kernelchaos"
KUBECTL = ["/usr/local/bin/kubectl", "--kubeconfig=/etc/kubernetes/admin.conf"]
CRI = ["/usr/local/bin/crictl", "--runtime-endpoint=unix:///run/containerd/containerd.sock",
       "--image-endpoint=unix:///run/containerd/containerd.sock", "--timeout=3s"]
CTR = ["/usr/local/bin/ctr", "--address=/run/containerd/containerd.sock", "--namespace=k8s.io"]
FIXTURE = "/usr/local/bin/cm-lifecycle-fixture"
RUNTIME = Path("/run/cm-lifecycle")
ROLES = ("target", "control", "bpfki", "daemon", "controller")
CONTAINER_NAMES = {"target": "worker", "control": "worker", "bpfki": "bpfki",
                   "daemon": "chaos-daemon", "controller": "chaos-mesh"}
IMAGE_ROLES = {"target": "worker", "control": "worker", "bpfki": "kernel",
               "daemon": "daemon", "controller": "controller"}
FINALIZER = "chaos-mesh/records"
MAX_EVIDENCE = 4 * 1024 * 1024


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def now():
    return time.clock_gettime(time.CLOCK_BOOTTIME)


def field(context, name):
    return context[name] if isinstance(context, dict) else getattr(context, name)


def remaining(deadline):
    value = deadline - now()
    require(value > 0, "stage or guest deadline expired")
    return value


def sleep_until(target, deadline):
    while True:
        current = now()
        require(current < deadline, "stage or guest deadline expired")
        delay = min(target, deadline) - current
        if delay <= 0:
            return
        time.sleep(min(0.1, delay))


def command(argv, deadline, data=None, missing=False, limit=512 * 1024):
    """Bound output and time for the fixture's fixed tool invocations."""
    timeout = min(3.0, remaining(deadline))
    process = subprocess.Popen(argv, stdin=subprocess.PIPE if data is not None else subprocess.DEVNULL,
                               stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    output = {process.stdout.fileno(): bytearray(), process.stderr.fileno(): bytearray()}
    selector = selectors.DefaultSelector()
    for stream in (process.stdout, process.stderr):
        os.set_blocking(stream.fileno(), False)
        selector.register(stream, selectors.EVENT_READ)
    try:
        if data is not None:
            require(len(data) < 16384, "tool input exceeded fixed bound")
            process.stdin.write(data)
            process.stdin.close()
        end = min(deadline, now() + timeout)
        while selector.get_map():
            for key, _ in selector.select(min(0.1, remaining(end))):
                chunk = os.read(key.fd, 65536)
                if not chunk:
                    selector.unregister(key.fileobj)
                    continue
                output[key.fd].extend(chunk)
                require(sum(map(len, output.values())) <= limit, "tool output exceeded bound")
        code = process.wait(timeout=remaining(end))
        stdout = bytes(output[process.stdout.fileno()])
        stderr = bytes(output[process.stderr.fileno()])
        if missing and code and b"(NotFound)" in stderr:
            return None
        require(code == 0, f"{Path(argv[0]).name} failed ({code}): {stderr.decode(errors='replace')[:2048]}")
        return stdout
    finally:
        selector.close()
        if process.poll() is None:
            process.kill()
            process.wait(timeout=1)
        process.stdout.close()
        process.stderr.close()


def kube(args, deadline, data=None, missing=False):
    duration = max(1, math.ceil(min(3, remaining(deadline))))
    result = command(KUBECTL + [f"--request-timeout={duration}s"] + args,
                     deadline, data, missing)
    return None if result is None else json.loads(result)


def read_process(pid):
    stat = Path(f"/proc/{pid}/stat").read_text()
    end = stat.rfind(")")
    require(end > 0, "invalid proc stat")
    start = int(stat[end + 2:].split()[19])
    status = Path(f"/proc/{pid}/status").read_text()
    line = next((line for line in status.splitlines() if line.startswith("NSpid:")), None)
    require(line is not None, "missing process NSpid")
    cgroup = Path(f"/proc/{pid}/cgroup").read_text().splitlines()
    require(len(cgroup) == 1 and cgroup[0].startswith("0::/"), "invalid unified process cgroup")
    return {"pid": pid, "starttime": start, "pid_namespace": os.stat(f"/proc/{pid}/ns/pid").st_ino,
            "nspid": [int(value) for value in line.split()[1:]], "cgroup": cgroup[0][3:]}


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


class Evidence:
    def __init__(self, directory):
        self.path = Path(directory) / "lifecycle-events.jsonl"
        self.stream = self.path.open("xb")
        self.size = 0
        self.sequence = 0
        self.pending_case = None

    def add(self, kind, value, case_uid=None):
        self.sequence += 1
        row = {"event": self.sequence, "kind": kind, "boottime_ns": time.clock_gettime_ns(time.CLOCK_BOOTTIME),
               "realtime_ns": time.time_ns(), "case_uid": case_uid, "value": value}
        data = (json.dumps(row, separators=(",", ":"), sort_keys=True) + "\n").encode()
        require(self.size + len(data) <= MAX_EVIDENCE, "required lifecycle evidence exceeds 4 MiB")
        if self.pending_case is not None and case_uid is None:
            self.pending_case.append(row)
        else:
            self.stream.write(data)
            self.stream.flush()
        self.size += len(data)
        return self.sequence

    def bind_case(self, uid):
        pending, self.pending_case = self.pending_case, None
        for row in pending:
            row["case_uid"] = uid
            data = (json.dumps(row, separators=(",", ":"), sort_keys=True) + "\n").encode()
            self.size += len(json.dumps(uid)) - 4
            require(self.size <= MAX_EVIDENCE, "required lifecycle evidence exceeds 4 MiB")
            self.stream.write(data)
        self.stream.flush()


class Observer:
    def __init__(self, context, evidence):
        self.context = context
        self.evidence = evidence
        self.identities = {}
        self.pidfds = {}
        self.process = None
        self.buffer = bytearray()
        self.case_uid = None
        self.watches = []
        self.cleanup_result = None
        self.counts = {"target": 0, "control": 0}
        self.deadline = field(context, "deadline")

    def discover(self, deadline):
        pods = select_pods(kube(["get", "pods", "--all-namespaces", "-o", "json"], deadline)["items"])
        tasks = command(CTR + ["tasks", "list"], deadline).decode().splitlines()
        require(tasks and tasks[0].split()[:3] == ["TASK", "PID", "STATUS"], "unexpected ctr task listing")
        task_pids = {}
        for line in tasks[1:]:
            columns = line.split()
            require(len(columns) == 3, "malformed ctr task row")
            if columns[2] == "RUNNING":
                task_pids[columns[0]] = int(columns[1])
        images = {entry["role"]: entry for entry in field(self.context, "bundle")["images"]}
        initial_ns = os.stat("/proc/1/ns/pid").st_ino
        for service in ("containerd.service", "kubelet.service"):
            pid = int(command(["/usr/bin/systemctl", "show", service, "--property=MainPID", "--value"], deadline))
            require(pid > 1 and os.stat(f"/proc/{pid}/ns/pid").st_ino == initial_ns,
                    f"{service} is not in initial PID namespace")
        for role in ROLES:
            identity = pod_identity(pods[role], role)
            container_id = identity["runtime_id"].removeprefix("containerd://")
            inspect = json.loads(command(CRI + ["inspect", "--output=json", container_id], deadline))
            pid = inspect["info"]["pid"]
            require(type(pid) is int and pid > 1 and task_pids.get(container_id) == pid,
                    "CRI PID differs from direct containerd task PID")
            validate_cri_identity(identity, inspect["status"])
            pod_slice = "kubepods-burstable-pod" + identity["pod_uid"].replace("-", "_") + ".slice"
            require(inspect["info"]["runtimeSpec"]["linux"]["cgroupsPath"]
                    == pod_slice + ":cri-containerd:" + container_id, "OCI cgroup differs from selected container")
            cgroup = "/kubepods.slice/kubepods-burstable.slice/" + pod_slice + "/cri-containerd-" + container_id + ".scope"
            process_identity = read_process(pid)
            pidfd = os.pidfd_open(pid, 0)
            try:
                require(process_identity["cgroup"] == cgroup, "process cgroup differs from selected container")
                if role in ("target", "control"):
                    require(process_identity["nspid"] == [pid, 1] and process_identity["pid_namespace"] != initial_ns,
                            "worker is not a separate container PID 1")
                if role in ("bpfki", "daemon"):
                    require(process_identity["nspid"] == [pid] and process_identity["pid_namespace"] == initial_ns,
                            "bpfki/daemon is not in initial PID namespace")
                require(read_process(pid) == process_identity and not select.select([pidfd], [], [], 0)[0],
                        "process changed while acquiring pidfd")
            except BaseException:
                os.close(pidfd)
                raise
            self.pidfds[role] = pidfd
            self.identities[role] = {**identity, **process_identity}
            image = json.loads(command(CRI + ["inspecti", "--output=json", inspect["status"]["imageRef"]], deadline))
            binding = image_binding(identity, inspect, image, images[IMAGE_ROLES[role]])
            self.identities[role]["image_binding"] = binding
            if role in ("target", "control"):
                runtime_spec = inspect["info"]["runtimeSpec"]
                security = runtime_spec["process"]
                require(runtime_spec["root"].get("readonly") is True, "worker root filesystem is writable")
                require(pods[role]["spec"].get("restartPolicy") == "Never" and not pods[role]["spec"].get("hostPID")
                        and not pods[role]["spec"].get("shareProcessNamespace"), "worker Pod isolation differs from fixture")
                state_path = "/var/lib/cm-lifecycle/" + role
                mounts = [mount for mount in runtime_spec["mounts"] if mount["destination"].startswith("/var/lib/cm-lifecycle")]
                require(len(mounts) == 1 and mounts[0]["destination"] == state_path and mounts[0]["source"] == state_path,
                        "worker state mount differs from its exclusive guest directory")
                require(security["user"]["uid"] == 10001 and security["user"]["gid"] == 10001,
                        "worker is not UID/GID 10001")
                require(security.get("noNewPrivileges") is True and not any(security.get("capabilities", {}).values()),
                        "worker privileges differ from fixed fixture")
                require(security["args"] == ["/cm-lifecycle-fixture", "worker", role], "worker has a wrapper or wrong args")
        require(len({entry["pid"] for entry in self.identities.values()}) == 5, "duplicate owned process")
        kernel = field(self.context, "bundle")["kernel"]
        release = kernel["release"]
        require(release == "6.8.0-cm-kernel-ci" and os.uname().release == release, "unexpected guest kernel release")
        root = Path(f"/proc/{self.identities['bpfki']['pid']}/root")
        headers = root / "usr/src" / ("linux-" + release)
        require(hashlib.sha256((headers / ".config").read_bytes()).hexdigest() == kernel["config_sha256"],
                "sidecar kernel config differs from sealed kernel")
        require((headers / "include/generated/utsrelease.h").read_text().strip() == f'#define UTS_RELEASE "{release}"',
                "sidecar generated release differs from running kernel")
        for name in ("build", "source"):
            require(os.readlink(root / "lib/modules" / release / name) == "/usr/src/linux-" + release,
                    "sidecar module header link differs from fixed header mount")
        boot_id = Path("/proc/sys/kernel/random/boot_id").read_text().strip()
        manifest = {"schema_version": 1, "boot_id": boot_id, "initial_pid_namespace": initial_ns,
                    "containers": [{"role": role, **{key: self.identities[role][key]
                                                     for key in ("pid", "starttime", "pid_namespace", "nspid")}}
                                   for role in ROLES]}
        descriptor = os.open(RUNTIME / "owned.json", os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_NOFOLLOW, 0o600)
        with os.fdopen(descriptor, "w") as output:
            output.write(json.dumps(manifest, separators=(",", ":")) + "\n")
        self.evidence.add("owned_containers", {"boot_id": boot_id, "containers": self.identities})
        self.process = subprocess.Popen([FIXTURE, "observe"], stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                        stderr=subprocess.PIPE, bufsize=0)
        os.set_blocking(self.process.stdout.fileno(), False)
        ready = self.read_response(min(deadline, now() + 3))
        require(ready.get("ok") and ready.get("command") == "ready", "native observer failed readiness")
        require(ready["header_sha256"] == field(self.context, "bundle")["worker"]["filter_header_sha256"],
                "observer filter header differs from bundle")
        self.evidence.add("observer_ready", ready)
        self.check_pods(deadline)

    def read_response(self, deadline):
        while b"\n" not in self.buffer:
            require(len(self.buffer) <= 16384, "observer response exceeded bound")
            require(self.process.poll() is None, "native observer exited")
            readable, _, _ = select.select([self.process.stdout], [], [], min(0.1, remaining(deadline)))
            if readable:
                chunk = os.read(self.process.stdout.fileno(), 16384)
                require(chunk, "native observer closed output")
                self.buffer.extend(chunk)
        line, _, rest = self.buffer.partition(b"\n")
        self.buffer = bytearray(rest)
        return json.loads(line)

    def check_pods(self, deadline):
        pods = select_pods(kube(["get", "pods", "--all-namespaces", "-o", "json"], deadline)["items"])
        for role in ROLES:
            current = pod_identity(pods[role], role)
            require(all(self.identities[role][key] == value for key, value in current.items()),
                    f"{role} Kubernetes identity changed")
            pidfd = self.pidfds[role]
            require(not select.select([pidfd], [], [], 0)[0], f"{role} process exited")
        return self.evidence.add("kubernetes_identity_check", {role: pod_identity(pods[role], role) for role in ROLES},
                                 self.case_uid)

    def ask(self, request, deadline):
        before = self.check_pods(deadline)
        self.process.stdin.write((request + "\n").encode())
        reply = self.read_response(min(deadline, now() + 3))
        self.evidence.add("native_sample", {"identity_before": before, "response": reply}, self.case_uid)
        require(reply.get("ok") is True and reply.get("command") == request, "native observation failed: " + str(reply))
        self.check_pods(deadline)
        return reply

    def operation(self, role, fault, deadline):
        reply = self.ask("operate " + role, deadline)
        self.counts[role] += 1
        require(reply["seq"] == self.counts[role], "worker sequence is not contiguous")
        require(reply["close_errno"] == 0 and reply["finished_ns"] - reply["started_ns"] <= 3_000_000_000,
                "worker operation deadline or close failed")
        require((reply["result"], reply["errno"]) == ((-1, errno.ENOMEM) if fault else (0, 0)),
                f"{role} operation did not {'fail with ENOMEM' if fault else 'succeed'}")
        return reply


class Watch:
    def __init__(self, observer, resource_version, deadline):
        require(re.fullmatch(r"[0-9]+", resource_version), "invalid known resourceVersion")
        self.observer = observer
        self.deadline = deadline
        self.buffer = bytearray()
        self.stderr = bytearray()
        self.deleted = False
        self.recovered = False
        self.closed = False
        query = urlencode({"watch": "true", "fieldSelector": "metadata.name=delete",
                           "resourceVersion": resource_version, "allowWatchBookmarks": "true",
                           "timeoutSeconds": max(1, math.ceil(remaining(deadline)))})
        self.process = subprocess.Popen(KUBECTL + [f"--request-timeout={max(1, math.ceil(remaining(deadline)))}s",
                                                   "get", "--raw", API + "?" + query],
                                        stdin=subprocess.DEVNULL, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        self.selector = selectors.DefaultSelector()
        for stream in (self.process.stdout, self.process.stderr):
            os.set_blocking(stream.fileno(), False)
            self.selector.register(stream, selectors.EVENT_READ)
        observer.watches.append(self)
        observer.evidence.add("watch_started", {"resource_version": resource_version}, observer.case_uid)

    def poll(self, wait=0):
        require(not self.closed, "watch is closed")
        for key, _ in self.selector.select(min(wait, remaining(self.deadline))):
            chunk = os.read(key.fd, 65536)
            if not chunk:
                self.selector.unregister(key.fileobj)
                continue
            if key.fileobj is self.process.stderr:
                self.stderr.extend(chunk)
                require(len(self.stderr) <= 8192, "watch error output exceeded bound")
                continue
            self.buffer.extend(chunk)
            require(len(self.buffer) <= 256 * 1024, "watch object exceeded bound")
            while b"\n" in self.buffer:
                line, _, rest = self.buffer.partition(b"\n")
                self.buffer = bytearray(rest)
                event = json.loads(line)
                self.observer.evidence.add("watch_event", event, self.observer.case_uid)
                kind, obj = event["type"], event["object"]
                require(kind != "ERROR", "watch history lost or API error: " + json.dumps(obj))
                require(kind in ("ADDED", "MODIFIED", "DELETED", "BOOKMARK"), "unknown watch event")
                if kind == "BOOKMARK":
                    continue
                require(obj["metadata"]["uid"] == self.observer.case_uid, "watch observed replacement UID")
                validate_object(obj, "delete", self.observer.case_uid)
                if recovered(obj):
                    self.recovered = True
                if kind == "DELETED":
                    require(obj["metadata"].get("deletionTimestamp"), "terminal object lacks deletionTimestamp")
                    self.deleted = True
        require(not self.stderr, "watch transport error: " + self.stderr.decode(errors="replace"))
        require(self.process.poll() is None or self.deleted, "watch ended before terminal deletion evidence")

    def close(self):
        if self.closed:
            return
        self.closed = True
        self.selector.close()
        if self.process.poll() is None:
            self.process.terminate()
            try:
                self.process.wait(timeout=0.5)
            except subprocess.TimeoutExpired:
                self.process.kill()
                self.process.wait(timeout=0.5)
        self.process.stdout.close()
        self.process.stderr.close()


def spec(kind):
    require(kind in ("expiration", "delete"), "invalid fixed case")
    result = {"mode": "one", "selector": {"pods": {NAMESPACE: ["target"]}},
              "containerNames": ["worker"], "failKernRequest": {
                  "failtype": 0, "probability": 100, "times": 0, "callchain": [{"predicate": "true"}]}}
    if kind == "expiration":
        result["duration"] = "120s"
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


def get_object(observer, kind, deadline, uid):
    obj = kube(["get", "--raw", API + "/" + kind], deadline)
    observer.evidence.add("object", obj, uid)
    validate_object(obj, kind, uid)
    return obj


def await_object(observer, kind, uid, deadline, predicate):
    while True:
        obj = get_object(observer, kind, deadline, uid)
        if predicate(obj):
            return obj
        sleep_until(now() + 0.25, deadline)


def clock_pair():
    before = time.clock_gettime_ns(time.CLOCK_BOOTTIME)
    real = time.time_ns()
    after = time.clock_gettime_ns(time.CLOCK_BOOTTIME)
    require(after - before < 10_000_000, "clock sampling was delayed")
    return {"boottime_ns": (before + after) // 2, "realtime_ns": real}


def check_clock(anchor):
    sample = clock_pair()
    drift = (sample["realtime_ns"] - anchor["realtime_ns"]) - (sample["boottime_ns"] - anchor["boottime_ns"])
    require(abs(drift) <= 100_000_000, "realtime/boottime clock discontinuity")
    return sample


def module_invariant():
    require(Path("/proc/modules").read_text().strip() == "", "guest has loaded modules")
    require(Path("/proc/sys/kernel/modules_disabled").read_text().strip() == "1", "module loading is unlocked")
    require(Path("/proc/sys/kernel/modprobe").read_text().strip() == "", "module autoload helper is enabled")


def fault_pair(observer, deadline):
    before = observer.ask("map active", deadline)
    target = observer.operation("target", True, deadline)
    after = observer.ask("map active", deadline)
    require(before["map_id"] == after["map_id"] and after["inject_times"] > before["inject_times"],
            "target ENOMEM did not increase owned filter counter")
    return target


def run_case(observer, kind):
    absolute = field(observer.context, "deadline")
    budget = 150 if kind == "expiration" else 90
    stage_deadline = min(absolute, now() + budget)
    module_invariant()
    observer.evidence.pending_case = []
    observer.ask("check", stage_deadline)
    observer.ask("map absent", stage_deadline)
    observer.operation("target", False, stage_deadline)
    observer.operation("control", False, stage_deadline)
    anchor = clock_pair()
    obj = kube(["create", "-f", "-", "-o", "json"], stage_deadline,
               json.dumps({"apiVersion": "chaos-mesh.org/v1alpha1", "kind": "KernelChaos",
                           "metadata": {"name": kind, "namespace": NAMESPACE}, "spec": spec(kind)}).encode())
    uid = obj["metadata"]["uid"]
    observer.case_uid = uid
    observer.evidence.bind_case(uid)
    observer.evidence.add("created", obj, uid)
    validate_object(obj, kind, uid)
    creation = datetime.datetime.fromisoformat(obj["metadata"]["creationTimestamp"].replace("Z", "+00:00")).timestamp()
    created_boot = anchor["boottime_ns"] / 1e9 + creation - anchor["realtime_ns"] / 1e9
    require(created_boot <= now() + 0.1 and now() - created_boot < 5, "creation timestamp is outside request window")
    stage_deadline = min(stage_deadline, created_boot + budget)
    observer.deadline = stage_deadline
    injection_deadline = min(stage_deadline, created_boot + 60)
    observer.evidence.add("case_deadlines", {"kind": kind, "created_boot": created_boot,
                                             "deadline": stage_deadline, "clock_anchor": anchor}, uid)
    await_object(observer, kind, uid, injection_deadline, active)
    observer.operation("control", False, injection_deadline)
    fault_pair(observer, injection_deadline)
    require(active(get_object(observer, kind, injection_deadline, uid)), "experiment is no longer active after first fault")
    if kind == "expiration":
        expiry = created_boot + 120
        sleep_until(expiry - 10, stage_deadline)
        check_clock(anchor)
        observer.operation("control", False, min(stage_deadline, expiry - 6))
        sleep_until(expiry - 6, stage_deadline)
        final_fault = fault_pair(observer, min(stage_deadline, expiry))
        finish = final_fault["finished_ns"] / 1e9
        require(expiry - 7 <= finish < expiry, "final fault missed last seven seconds before expiration")
        require(active(get_object(observer, kind, min(stage_deadline, expiry), uid)), "expiration recovered before final active evidence")
        observer.evidence.add("pre_expiry_fault", {"expiry_boottime": expiry, "operation": final_fault,
                                                   "clock": check_clock(anchor)}, uid)
        sleep_until(expiry, stage_deadline)
        await_object(observer, kind, uid, stage_deadline, expiration_recovered)
        observer.evidence.add("expiration_clock", check_clock(anchor), uid)
    else:
        known = get_object(observer, kind, stage_deadline, uid)
        require(active(known), "Delete watch must begin from active known object")
        watch = Watch(observer, known["metadata"]["resourceVersion"], stage_deadline)
        observer.operation("control", False, stage_deadline)
        final_fault = fault_pair(observer, stage_deadline)
        trigger_deadline = min(stage_deadline, final_fault["finished_ns"] / 1e9 + 5)
        require(active(get_object(observer, kind, trigger_deadline, uid)), "experiment is not active before Delete")
        watch.poll()
        options = {"apiVersion": "v1", "kind": "DeleteOptions", "preconditions": {"uid": uid}}
        options_file = RUNTIME / "delete-options.json"
        options_file.write_text(json.dumps(options))
        accepted = kube(["delete", "--raw", API + "/delete", "-f", str(options_file)], trigger_deadline)
        accepted_at = now()
        require(accepted_at <= trigger_deadline, "UID Delete completed more than five seconds after final fault")
        observer.evidence.add("delete_accepted", {"response": accepted, "accepted_boottime": accepted_at}, uid)
        recovery_deadline = min(stage_deadline, accepted_at + 30)
        watch.deadline = recovery_deadline
        while not watch.deleted:
            watch.poll(0.1)
        require(watch.recovered and not watch.buffer, "Delete lacks complete successful recovery watch evidence")
        absent = kube(["get", "--raw", API + "/delete"], recovery_deadline, missing=True)
        require(absent is None, "deleted object still exists")
        observer.evidence.add("deleted_get", {"http_status": 404, "uid": uid}, uid)
        watch.close()
        stage_deadline = recovery_deadline
    observer.ask("map recovered", stage_deadline)
    observer.operation("target", False, stage_deadline)
    observer.operation("control", False, stage_deadline)
    observer.ask("check", stage_deadline)
    module_invariant()
    return {"kind": kind, "uid": uid, "ok": True, "created_at": obj["metadata"]["creationTimestamp"],
            "deadline": stage_deadline, "finished_boottime": now()}


def cleanup(context, observer):
    if observer is None:
        observer = context.get("_lifecycle_observer") if isinstance(context, dict) else getattr(context, "_lifecycle_observer", None)
    saved_deadline = context.get("cleanup_deadline") if isinstance(context, dict) else getattr(context, "cleanup_deadline", None)
    deadline = min(field(context, "deadline"), saved_deadline if saved_deadline is not None else now() + 30)
    if isinstance(context, dict):
        context["cleanup_deadline"] = deadline
    else:
        setattr(context, "cleanup_deadline", deadline)
    if observer is not None and observer.cleanup_result is not None:
        return observer.cleanup_result
    result = {"ok": False, "events": [], "errors": []}
    stopped = False
    try:
        command(["/usr/bin/systemctl", "stop", "kubelet.service"], deadline)
        state = command(["/usr/bin/systemctl", "show", "kubelet.service", "--property=ActiveState", "--value"], deadline).decode().strip()
        require(state == "inactive", "kubelet supervision did not stop")
        stopped = True
        result["events"].append({"action": "stop_kubelet", "state": state})
    except Exception as error:
        result["errors"].append(str(error))
    if observer is not None:
        for watch in observer.watches:
            try:
                watch.close()
            except Exception as error:
                result["errors"].append("watch cleanup: " + str(error))
        if stopped:
            for role in ("bpfki", "target", "control"):
                if role not in observer.pidfds:
                    continue
                pidfd = observer.pidfds[role]
                try:
                    exited = bool(select.select([pidfd], [], [], 0)[0])
                    if not exited:
                        signal.pidfd_send_signal(pidfd, signal.SIGKILL)
                        require(bool(select.select([pidfd], [], [], min(3, remaining(deadline)))[0]),
                                f"{role} did not exit after pidfd signal")
                    result["events"].append({"action": "owned_process_stopped", "role": role,
                                             "pid": observer.identities[role]["pid"], "already_exited": exited})
                except Exception as error:
                    result["errors"].append(f"{role} cleanup: {error}")
                    if role == "bpfki":
                        break
        if observer.process is not None:
            try:
                if observer.process.poll() is None:
                    observer.process.stdin.write(b"quit\n")
                    observer.process.wait(timeout=min(1, remaining(deadline)))
                native_error = observer.process.stderr.read(8193)
                require(len(native_error) <= 8192, "native stderr exceeded bound")
                require(observer.process.returncode == 0, "native observer exited abnormally: " + native_error.decode(errors="replace"))
            except Exception as error:
                result["errors"].append("observer shutdown: " + str(error))
                if observer.process.poll() is None:
                    observer.process.kill()
                    observer.process.wait(timeout=1)
            finally:
                for stream in (observer.process.stdin, observer.process.stdout, observer.process.stderr):
                    stream.close()
        for role, pidfd in list(observer.pidfds.items()):
            try:
                os.close(pidfd)
            except OSError as error:
                result["errors"].append(f"{role} pidfd close: {error}")
            del observer.pidfds[role]
        try:
            module_invariant()
        except Exception as error:
            result["errors"].append(str(error))
        observer.cleanup_result = result
    result["ok"] = stopped and not result["errors"]
    if observer is not None:
        try:
            observer.evidence.add("cleanup", result)
        except Exception as error:
            result["ok"] = False
            result["errors"].append("cleanup evidence: " + str(error))
        observer.evidence.stream.close()
    return result


def run(context):
    result = {"schema_version": 1, "ok": False, "cases": [], "error": None, "cleanup": None}
    observer = None
    try:
        require(os.geteuid() == 0 and Path("/proc/sys/kernel/random/boot_id").is_file(), "verifier requires guest root")
        bundle = field(context, "bundle")
        result.update({"source_sha": bundle["source_sha"], "bpfki_commit": bundle["bpfki_commit"],
                       "kernel": bundle["kernel"], "worker": bundle["worker"]})
        actual_binary = hashlib.sha256(Path(FIXTURE).read_bytes()).hexdigest()
        require(actual_binary == bundle["worker"]["binary_sha256"], "worker binary differs from sealed input")
        evidence = Evidence(field(context, "evidence_dir"))
        observer = Observer(context, evidence)
        if isinstance(context, dict):
            context["_lifecycle_observer"] = observer
        else:
            setattr(context, "_lifecycle_observer", observer)
        supplied = context.get("fixture_deadline") if isinstance(context, dict) else getattr(context, "fixture_deadline", None)
        fixture_deadline = min(field(context, "deadline"), supplied if supplied is not None else now() + 30)
        observer.discover(fixture_deadline)
        initial = kube(["get", "--raw", API], fixture_deadline)
        require(initial["items"] == [], "fixture namespace contains existing KernelChaos")
        for kind in ("expiration", "delete"):
            observer.case_uid = None
            result["cases"].append(run_case(observer, kind))
        expiration = result["cases"][0]
        final_deadline = result["cases"][-1]["deadline"]
        require(expiration_recovered(get_object(observer, "expiration", final_deadline, expiration["uid"])),
                "expired object did not remain stopped and present through Delete case")
        require(observer.counts == {"target": 8, "control": 8}, "unexpected worker operation count")
        remaining(final_deadline)
        result["ok"] = True
        result["evidence"] = str(evidence.path)
        evidence.add("lifecycle_result", result)
        return result, observer
    except BaseException as error:
        result["ok"] = False
        result["error"] = f"{type(error).__name__}: {error}"
        if observer is not None:
            try:
                if observer.evidence.pending_case is not None:
                    observer.evidence.bind_case(observer.case_uid)
                observer.evidence.add("lifecycle_failure", result)
            except Exception as evidence_error:
                result["error"] += "; evidence failure: " + str(evidence_error)
        try:
            result["cleanup"] = cleanup(context, observer)
        except BaseException as cleanup_error:
            result["cleanup"] = {"ok": False, "events": [], "errors": [str(cleanup_error)]}
        return result, observer
