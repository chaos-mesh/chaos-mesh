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
from urllib.parse import urlencode


from checks import (API, ROLES, IMAGE_ROLES, require,
                    pod_identity, select_pods, validate_cri_identity, image_binding, validate_object, recovered)
from cases import Evidence, Observer as FixedObserver, now, field, remaining, run_cases, command, read_process

KUBECTL = ["/usr/local/bin/kubectl", "--kubeconfig=/etc/kubernetes/admin.conf"]
CRI = ["/usr/local/bin/crictl", "--runtime-endpoint=unix:///run/containerd/containerd.sock",
       "--image-endpoint=unix:///run/containerd/containerd.sock", "--timeout=3s"]
CTR = ["/usr/local/bin/ctr", "--address=/run/containerd/containerd.sock", "--namespace=k8s.io"]
FIXTURE = "/usr/local/bin/cm-lifecycle-fixture"
RUNTIME = Path("/run/cm-lifecycle")


def kube(args, deadline, data=None, missing=False):
    duration = max(1, math.ceil(min(15, remaining(deadline))))
    result = command(KUBECTL + [f"--request-timeout={duration}s"] + args,
                     deadline, data, missing, timeout=15)
    return None if result is None else json.loads(result)


class Observer(FixedObserver):
    def __init__(self, context, evidence):
        super().__init__(context, evidence)
        self.api = SealedAPI()
        self.invariant = module_invariant

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


def module_invariant():
    require(Path("/proc/modules").read_text().strip() == "", "guest has loaded modules")
    require(Path("/proc/sys/kernel/modules_disabled").read_text().strip() == "1", "module loading is unlocked")
    require(Path("/proc/sys/kernel/modprobe").read_text().strip() == "", "module autoload helper is enabled")


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
        result["cases"] = run_cases(observer)
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

class SealedAPI:
    def pods(self, deadline):
        return kube(["get", "pods", "--all-namespaces", "-o", "json"], deadline)

    def get(self, name, deadline, missing=False):
        return kube(["get", "--raw", API + "/" + name], deadline, missing=missing)

    def create(self, obj, deadline):
        return kube(["create", "-f", "-", "-o", "json"], deadline, json.dumps(obj).encode())

    def delete(self, name, uid, deadline):
        options = {"apiVersion": "v1", "kind": "DeleteOptions", "preconditions": {"uid": uid}}
        path = RUNTIME / "delete-options.json"
        path.write_text(json.dumps(options))
        return kube(["delete", "--raw", API + "/" + name, "-f", str(path)], deadline)

    def watch(self, observer, resource_version, deadline):
        return Watch(observer, resource_version, deadline)
