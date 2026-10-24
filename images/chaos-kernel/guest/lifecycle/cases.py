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

import datetime
import errno
import json
import os
from pathlib import Path
import select
import selectors
import subprocess
import time
from checks import ( NAMESPACE, ROLES, BUDGET, require, pod_identity, select_pods,
                    spec, validate_object, active, expiration_recovered)

MAX_EVIDENCE = 4 * 1024 * 1024


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
        self.faults = 0

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
        pods = select_pods(self.api.pods(deadline)["items"])
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
        remaining(deadline)
        self.process.stdin.write((request + "\n").encode())
        reply = self.read_response(min(deadline, now() + BUDGET.operation))
        self.evidence.add("native_sample", {"identity_before": before, "response": reply}, self.case_uid)
        require(reply.get("ok") is True and reply.get("command") == request, "native observation failed: " + str(reply))
        self.check_pods(deadline)
        return reply

    def operation(self, role, fault, deadline):
        require(self.counts[role] < BUDGET.operations, "worker operation limit reached")
        require(not fault or (role == "target" and self.faults < BUDGET.faults), "fault operation limit reached")
        reply = self.ask("operate " + role, deadline)
        self.faults += int(fault)
        self.counts[role] += 1
        require(reply["seq"] == self.counts[role], "worker sequence is not contiguous")
        require(reply["close_errno"] == 0 and reply["finished_ns"] - reply["started_ns"] <= BUDGET.operation * 1_000_000_000,
                "worker operation deadline or close failed")
        require((reply["result"], reply["errno"]) == ((-1, errno.ENOMEM) if fault else (0, 0)),
                f"{role} operation did not {'fail with ENOMEM' if fault else 'succeed'}")
        return reply


def get_object(observer, kind, deadline, uid):
    obj = observer.api.get(kind, deadline)
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


def fault_pair(observer, deadline):
    before = observer.ask("map active", deadline)
    target = observer.operation("target", True, deadline)
    after = observer.ask("map active", deadline)
    require(before["map_id"] == after["map_id"] and after["inject_times"] > before["inject_times"],
            "target ENOMEM did not increase owned filter counter")
    return target


def run_case(observer, kind):
    absolute = field(observer.context, "deadline")
    budget = getattr(BUDGET, kind)
    stage_deadline = min(absolute, now() + budget)
    observer.invariant()
    observer.evidence.pending_case = []
    observer.ask("check", stage_deadline)
    observer.ask("map absent", stage_deadline)
    observer.operation("target", False, stage_deadline)
    observer.operation("control", False, stage_deadline)
    anchor = clock_pair()
    obj = observer.api.create({"apiVersion": "chaos-mesh.org/v1alpha1", "kind": "KernelChaos",
                           "metadata": {"name": kind, "namespace": NAMESPACE}, "spec": spec(kind)}, stage_deadline)
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
    injection_deadline = min(stage_deadline, created_boot + BUDGET.first_injection)
    observer.evidence.add("case_deadlines", {"kind": kind, "created_boot": created_boot,
                                             "deadline": stage_deadline, "clock_anchor": anchor}, uid)
    await_object(observer, kind, uid, injection_deadline, active)
    observer.operation("control", False, injection_deadline)
    fault_pair(observer, injection_deadline)
    require(active(get_object(observer, kind, injection_deadline, uid)), "experiment is no longer active after first fault")
    if kind == "expiration":
        expiry = created_boot + BUDGET.duration
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
        watch = observer.api.watch(observer, known["metadata"]["resourceVersion"], stage_deadline)
        observer.operation("control", False, stage_deadline)
        final_fault = fault_pair(observer, stage_deadline)
        trigger_deadline = min(stage_deadline, final_fault["finished_ns"] / 1e9 + 5)
        require(active(get_object(observer, kind, trigger_deadline, uid)), "experiment is not active before Delete")
        watch.poll()
        accepted = observer.api.delete("delete", uid, trigger_deadline)
        accepted_at = now()
        require(accepted_at <= trigger_deadline, "UID Delete completed more than five seconds after final fault")
        observer.evidence.add("delete_accepted", {"response": accepted, "accepted_boottime": accepted_at}, uid)
        recovery_deadline = min(stage_deadline, accepted_at + 30)
        watch.deadline = recovery_deadline
        while not watch.deleted:
            watch.poll(0.1)
        require(watch.recovered and not watch.buffer, "Delete lacks complete successful recovery watch evidence")
        absent = observer.api.get("delete", recovery_deadline, missing=True)
        require(absent is None, "deleted object still exists")
        observer.evidence.add("deleted_get", {"http_status": 404, "uid": uid}, uid)
        watch.close()
        stage_deadline = recovery_deadline
    observer.ask("map recovered", stage_deadline)
    observer.operation("target", False, stage_deadline)
    observer.operation("control", False, stage_deadline)
    observer.ask("check", stage_deadline)
    observer.invariant()
    return {"kind": kind, "uid": uid, "ok": True, "created_at": obj["metadata"]["creationTimestamp"],
            "deadline": stage_deadline, "finished_boottime": now()}


def run_cases(observer):
    results = []
    for kind in ("expiration", "delete"):
        observer.case_uid = None
        results.append(run_case(observer, kind))
    final_deadline = results[-1]["deadline"]
    require(expiration_recovered(get_object(observer, "expiration", final_deadline, results[0]["uid"])),
            "expired object did not remain stopped and present through Delete case")
    require(observer.counts == {"target": BUDGET.operations, "control": BUDGET.operations}
            and observer.faults == BUDGET.faults, "unexpected worker operation count")
    remaining(final_deadline)
    return results


def command(argv, deadline, data=None, missing=False, limit=512 * 1024, timeout=3):
    timeout = min(timeout, remaining(deadline))
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

