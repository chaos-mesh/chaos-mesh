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
#


import hashlib
import json
import os
from pathlib import Path
import select
import subprocess

from api import APIClient, APIError
from cases import Observer as FixedObserver, Evidence, command, now, field, read_process, run_cases
from checks import (API, NAMESPACE, ROLES, IMAGE_ROLES, BUDGET, require, pod_identity,
                    select_pods, validate_cri_identity, image_binding, spec)
from identity import Guest, require_gate, write_json

FIXTURE = "/tools/cm-lifecycle-fixture"
RUNTIME = Path("/run/cm-lifecycle")
CRI = ["/tools/crictl", "--runtime-endpoint=unix:///run/containerd/containerd.sock",
       "--image-endpoint=unix:///run/containerd/containerd.sock", "--timeout=3s"]
CTR = ["/tools/ctr", "--address=/run/containerd/containerd.sock", "--namespace=k8s.io"]


def live_guest(guest):
    guest.check_live(Path('/proc/sys/kernel/random/boot_id').read_text().strip(),
                     Path('/host-etc/machine-id').read_text().strip(), os.stat('/proc/1/ns/pid').st_ino,
                     os.uname().release, os.uname().machine)
    require(os.geteuid() == 0 and os.stat('/proc/self/ns/pid').st_ino == guest.initial_pid_namespace,
            'observer is not guest root in the initial PID namespace')


class Observer(FixedObserver):
    def __init__(self, context, evidence, api, guest):
        super().__init__(context, evidence)
        self.api, self.guest = api, guest

    def invariant(self):
        live_guest(self.guest)

    def discover(self, deadline):
        pods = select_pods(self.api.pods(deadline)["items"])
        tasks = command(CTR + ["tasks", "list"], deadline).decode().splitlines()
        require(tasks and tasks[0].split()[:3] == ["TASK", "PID", "STATUS"], "unexpected ctr task listing")
        task_pids = {}
        for line in tasks[1:]:
            columns = line.split()
            require(len(columns) == 3, "malformed ctr task row")
            if columns[2] == "RUNNING":
                task_pids[columns[0]] = int(columns[1])
        images = field(self.context, "bundle")["images"]
        initial_ns = os.stat("/proc/1/ns/pid").st_ino
        for role in ROLES:
            identity = pod_identity(pods[role], role)
            container_id = identity["runtime_id"].removeprefix("containerd://")
            inspect = json.loads(command(CRI + ["inspect", "--output=json", container_id], deadline))
            pid = inspect["info"]["pid"]
            require(type(pid) is int and pid > 1 and task_pids.get(container_id) == pid,
                    "CRI PID differs from direct containerd task PID")
            validate_cri_identity(identity, inspect["status"])
            cgroup_path = inspect["info"]["runtimeSpec"]["linux"]["cgroupsPath"]
            if cgroup_path.startswith("/"):
                cgroup = cgroup_path
            else:
                parts = cgroup_path.split(":")
                require(len(parts) == 3 and parts[1:] == ["cri-containerd", container_id], "unexpected OCI cgroup path")
                pod_slice = "kubepods-burstable-pod" + identity["pod_uid"].replace("-", "_") + ".slice"
                require(parts[0] == pod_slice, "OCI cgroup is not owned by selected Pod")
                cgroup = "/kubepods.slice/kubepods-burstable.slice/" + pod_slice + "/cri-containerd-" + container_id + ".scope"
            require(container_id in cgroup and (identity["pod_uid"] in cgroup or
                    identity["pod_uid"].replace("-", "_") in cgroup), "cgroup does not bind Pod and container")
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


def admission(api, evidence, deadline):
    obj = {'apiVersion': 'chaos-mesh.org/v1alpha1', 'kind': 'KernelChaos',
           'metadata': {'name': 'admission-check', 'namespace': NAMESPACE}, 'spec': spec('expiration')}
    path = API + '?dryRun=All&fieldValidation=Strict'
    obj['spec']['unknownQualificationField'] = True
    try:
        api.request('POST', path, deadline, obj)
    except APIError as error:
        require(error.status == 400 and 'unknown field' in error.value.get('message', ''), 'strict fields were not enforced')
        evidence.add('strict_field_rejection', error.value)
    else:
        raise RuntimeError('strict field dry-run accepted an unknown field')
    del obj['spec']['unknownQualificationField']
    obj['spec']['duration'] = 'not-a-duration'
    try:
        api.request('POST', path, deadline, obj)
    except APIError as error:
        message = error.value.get('message', '')
        require('admission webhook "vkernelchaos.kb.io"' in message and 'parse duration field error' in message,
                'invalid duration was not rejected by the expected webhook')
        evidence.add('admission_rejection', error.value)
    else:
        raise RuntimeError('invalid duration was accepted')
    obj['spec']['duration'] = '120s'
    value = api.request('POST', path, deadline, obj)
    require(value['spec']['selector']['namespaces'] == [NAMESPACE], 'namespace admission default missing')
    require(api.get('admission-check', deadline, missing=True) is None, 'dry-run persisted a resource')
    evidence.add('admission_success', value)


def run():
    inputs = json.loads(Path('/tools/inputs.json').read_text())
    guest = Guest.read(json.loads(Path('/etc/cm-lifecycle/guest.json').read_text()), inputs['profile'])
    live_guest(guest)
    for name, checksum in inputs['files'].items():
        require(hashlib.sha256(Path('/tools', name).read_bytes()).hexdigest() == checksum,
                'mounted executable or verifier changed: ' + name)
    require(hashlib.sha256(Path(FIXTURE).read_bytes()).hexdigest() == inputs['worker']['binary_sha256'], 'fixture changed')
    evidence = Evidence('/evidence')
    api = APIClient()
    observer = None
    result = {'ok': False, 'cases': [], 'cleanup': {'status': 'host-owned'}}
    try:
        deadline = now() + 90
        context = {'bundle': inputs, 'deadline': deadline, 'evidence_dir': '/evidence'}
        observer = Observer(context, evidence, api, guest)
        observer.discover(deadline)
        expected = {key: observer.identities['bpfki'][key] for key in
                    ('pod', 'pod_uid', 'runtime_id', 'image', 'pod_image_id')}
        receipt = json.loads(Path('/etc/cm-lifecycle/gate.json').read_text())
        require_gate(receipt, guest, expected)
        node = api.request('GET', '/api/v1/nodes/' + guest.profile, deadline)
        require(node['metadata']['uid'] == guest.node_uid and node['status']['nodeInfo']['bootID'] == guest.boot_id,
                'Kubernetes node identity differs from the host')
        admission(api, evidence, deadline)
        require(api.request('GET', API, deadline)['items'] == [], 'fixture namespace contains existing KernelChaos')
        write_json('/evidence/ready.json', {'guest': guest.__dict__, 'bpfki': expected})
        lease_path = Path('/etc/cm-lifecycle/active.json')
        while not lease_path.exists():
            from cases import sleep_until
            sleep_until(now() + 0.1, deadline)
        lease = json.loads(lease_path.read_text())
        require(lease['guest'] == guest.__dict__ and 0 < lease['deadline_boottime'] - now() <= BUDGET.active - 5,
                'active lease is missing, expired, or belongs to another guest')
        context['deadline'] = lease['deadline_boottime']
        observer.deadline = context['deadline']
        api.arm(receipt, guest, expected, lease)
        evidence.add('armed', lease)
        result['cases'] = run_cases(observer)
        result['ok'] = True
    except BaseException as error:
        result['error'] = type(error).__name__ + ': ' + str(error)
        print(result['error'], flush=True)
    finally:
        api.armed = False
        if observer is not None:
            for watch in observer.watches:
                watch.close()
            for kind in ('expiration', 'delete'):
                try:
                    obj = api.get(kind, now() + 3, missing=True)
                    if obj is not None:
                        api.delete(kind, obj['metadata']['uid'], now() + 3)
                except Exception as error:
                    result.setdefault('cleanup_errors', []).append(str(error))
            if observer.process is not None:
                observer.process.kill()
                observer.process.wait(timeout=1)
            for fd in observer.pidfds.values():
                os.close(fd)
        if evidence.pending_case is not None:
            evidence.bind_case(observer.case_uid if observer is not None else None)
        evidence.add('recovery_result', result)
        evidence.stream.close()
        write_json('/evidence/recovery.json', result)
    return 0 if result['ok'] else 1


if __name__ == '__main__':
    raise SystemExit(run())
