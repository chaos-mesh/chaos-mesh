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


"""Install exact inputs into the workflow-owned VM, then require qualification and recovery."""
import argparse
from dataclasses import asdict
import json
import os
from pathlib import Path
import shlex
import signal
import subprocess
import sys
import tarfile
import time

from identity import Guest, Workflow, need, require_gate, verify_inputs, write_json
import probe

REPO = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(REPO / 'images/chaos-kernel/guest/lifecycle'))
from checks import BUDGET, NAMESPACE, pod_identity, select_pods
from workload import chart_values, worker_pods


class Host:
    def __init__(self, profile, inputs, output):
        self.profile, self.inputs, self.output = profile, inputs, output
        self.count = 0
        self.kube = [str(inputs / 'kubectl'), '--kubeconfig=' + os.environ['KUBECONFIG'], '--context=' + profile]

    def run(self, args, seconds=30, data=None):
        self.count += 1
        path = self.output / f'command-{self.count:03}.json'
        with subprocess.Popen([str(arg) for arg in args], stdin=subprocess.PIPE if data is not None else subprocess.DEVNULL,
                              stdout=subprocess.PIPE, stderr=subprocess.PIPE, start_new_session=True) as process:
            try:
                stdout, stderr = process.communicate(data, timeout=seconds)
            except subprocess.TimeoutExpired:
                os.killpg(process.pid, signal.SIGKILL)
                stdout, stderr = process.communicate()
                path.write_text(json.dumps({'argv': [str(a) for a in args], 'timeout': seconds}))
                raise
            need(len(stdout) + len(stderr) <= 8 * 1024 * 1024, 'command evidence exceeded bound')
            path.write_text(json.dumps({'argv': [str(a) for a in args], 'returncode': process.returncode,
                                       'stdout': stdout.decode(errors='replace'),
                                       'stderr': stderr.decode(errors='replace')}, indent=2))
            need(process.returncode == 0, 'command failed: ' + str(path))
            return stdout

    def kubectl(self, *args, seconds=30, data=None):
        return self.run(self.kube + ['--request-timeout=' + str(seconds) + 's', *args], seconds + 2, data)

    def ssh(self, script, seconds=30):
        return self.run(['minikube', '-p', self.profile, 'ssh', '--', 'sudo sh -ec ' + shlex.quote(script)], seconds)

    def install_json(self, name, value):
        path = self.output / (name + '.json')
        write_json(path, value)
        self.run(['minikube', '-p', self.profile, 'cp', path, self.profile + ':/tmp/cm-' + name + '.json'])
        destination = '/var/lib/cm-minikube/identity/' + name + '.json'
        self.ssh('install -o root -g root -m 0444 /tmp/cm-' + name + '.json ' + destination + '.new; mv ' + destination + '.new ' + destination)

    def read_evidence(self, name):
        return self.ssh('cat /var/lib/cm-minikube/evidence/' + name, 10)


def tool_pod(image):
    mounts = [('tools', '/tools', '/var/lib/cm-minikube/tools', True, 'Directory'),
              ('identity', '/etc/cm-lifecycle', '/var/lib/cm-minikube/identity', True, 'Directory'),
              ('evidence', '/evidence', '/var/lib/cm-minikube/evidence', False, 'Directory'),
              ('state', '/var/lib/cm-lifecycle', '/var/lib/cm-lifecycle', False, 'Directory'),
              ('machine', '/host-etc/machine-id', '/etc/machine-id', True, 'File'),
              ('socket', '/run/containerd/containerd.sock', '/run/containerd/containerd.sock', True, 'Socket')]
    return {'apiVersion': 'v1', 'kind': 'Pod', 'metadata': {'name': 'observer', 'namespace': NAMESPACE},
            'spec': {'hostPID': True, 'restartPolicy': 'Never', 'serviceAccountName': 'observer',
                     'containers': [{'name': 'observer', 'image': image, 'imagePullPolicy': 'Never',
                                     'command': ['sleep', 'infinity'], 'securityContext': {'privileged': True},
                                     'resources': {'requests': {'cpu': '10m', 'memory': '64Mi'},
                                                   'limits': {'cpu': '500m', 'memory': '256Mi'}},
                                     'volumeMounts': [{'name': name, 'mountPath': dest, 'readOnly': readonly}
                                                      for name, dest, source, readonly, kind in mounts] +
                                                     [{'name': 'runtime', 'mountPath': '/run/cm-lifecycle'}]}],
                     'volumes': [{'name': name, 'hostPath': {'path': source, 'type': kind}}
                                 for name, dest, source, readonly, kind in mounts] +
                                [{'name': 'runtime', 'emptyDir': {}}]}}


def observer_rbac():
    return [
        {'apiVersion': 'v1', 'kind': 'ServiceAccount', 'metadata': {'name': 'observer', 'namespace': NAMESPACE}},
        {'apiVersion': 'rbac.authorization.k8s.io/v1', 'kind': 'ClusterRole', 'metadata': {'name': 'cm-kernel-observer'},
         'rules': [{'apiGroups': [''], 'resources': ['pods'], 'verbs': ['get', 'list']},
                   {'apiGroups': [''], 'resources': ['nodes'], 'verbs': ['get']}]},
        {'apiVersion': 'rbac.authorization.k8s.io/v1', 'kind': 'ClusterRoleBinding', 'metadata': {'name': 'cm-kernel-observer'},
         'roleRef': {'apiGroup': 'rbac.authorization.k8s.io', 'kind': 'ClusterRole', 'name': 'cm-kernel-observer'},
         'subjects': [{'kind': 'ServiceAccount', 'name': 'observer', 'namespace': NAMESPACE}]},
        {'apiVersion': 'rbac.authorization.k8s.io/v1', 'kind': 'Role', 'metadata': {'name': 'observer', 'namespace': NAMESPACE},
         'rules': [{'apiGroups': ['chaos-mesh.org'], 'resources': ['kernelchaos'], 'verbs': ['get', 'list', 'watch', 'create', 'delete']}]},
        {'apiVersion': 'rbac.authorization.k8s.io/v1', 'kind': 'RoleBinding', 'metadata': {'name': 'observer', 'namespace': NAMESPACE},
         'roleRef': {'apiGroup': 'rbac.authorization.k8s.io', 'kind': 'Role', 'name': 'observer'},
         'subjects': [{'kind': 'ServiceAccount', 'name': 'observer', 'namespace': NAMESPACE}]},
    ]


def wait_file(path, deadline, process=None):
    while not path.exists():
        need(time.monotonic() < deadline, 'timed out waiting for ' + path.name)
        need(process is None or process.poll() is None, 'supervised process exited before ' + path.name)
        time.sleep(0.1)
    return json.loads(path.read_text())


def qualify(profile, inputs, output):
    workflow = Workflow.current()
    need(profile == workflow.profile, 'profile differs from current run')
    bundle = verify_inputs(inputs / 'inputs.json', workflow)
    for name in ('kubectl', 'helm', 'crictl', 'ctr', 'cm-lifecycle-fixture'):
        (inputs / name).chmod(0o555)
    guest = Guest.read(json.loads((output / 'guest.json').read_text()), profile)
    host = Host(profile, inputs, output)
    supervisor = process = None
    result = {'recovery': {'ok': False}, 'containment': {'ok': False}}
    try:
        for entry in bundle['images'].values():
            host.run(['minikube', '-p', profile, 'image', 'load', inputs / entry['archive']], 180)
        host.ssh('test ! -e /var/lib/cm-minikube; test ! -e /var/lib/cm-lifecycle; '
                 'mkdir -p /var/lib/cm-minikube/tools /var/lib/cm-minikube/identity /var/lib/cm-minikube/evidence /var/lib/cm-lifecycle')
        bundle['profile'] = profile
        staged = output / 'inputs.json'
        staged.write_text(json.dumps(bundle, indent=2))
        with tarfile.open(output / 'tools.tar', 'w') as archive:
            archive.add(staged, arcname='inputs.json')
            for name in bundle['files']:
                archive.add(inputs / name, arcname=name)
        host.run(['minikube', '-p', profile, 'cp', output / 'tools.tar', profile + ':/tmp/cm-tools.tar'], 60)
        host.ssh('tar -xf /tmp/cm-tools.tar -C /var/lib/cm-minikube/tools; '
                 'chown -R root:root /var/lib/cm-minikube; chmod -R a-w /var/lib/cm-minikube/tools')
        host.install_json('guest', asdict(guest))
        identity_line = '\n'.join(str(getattr(guest, key)) for key in
                                  ('profile', 'domain_uuid', 'boot_id', 'node_uid', 'machine_id', 'initial_pid_namespace', 'nonce')) + '\n'
        host.ssh('printf %s ' + shlex.quote(identity_line) + ' > /var/lib/cm-minikube/identity/minikube-identity; '
                 'chmod 0444 /var/lib/cm-minikube/identity/minikube-identity')
        host.kubectl('create', 'namespace', NAMESPACE)
        resources = observer_rbac() + [tool_pod(bundle['images']['kernel']['tag'])]
        host.kubectl('create', '-f', '-', data=json.dumps({'apiVersion': 'v1', 'kind': 'List', 'items': resources}).encode())
        host.kubectl('wait', '--for=condition=Ready', 'pod/observer', '-n', NAMESPACE, '--timeout=90s', seconds=95)
        host.kubectl('exec', '-n', NAMESPACE, 'observer', '-c', 'observer', '--', '/tools/cm-lifecycle-fixture', 'prepare')
        host.kubectl('create', '-f', '-', data=json.dumps({'apiVersion': 'v1', 'kind': 'List',
                     'items': worker_pods(bundle['images']['worker']['tag'])}).encode())
        values = output / 'values.json'
        values.write_text(json.dumps(chart_values(bundle['images']), indent=2))
        host.run([inputs / 'helm', '--kubeconfig', os.environ['KUBECONFIG'], '--kube-context', profile,
                  'install', 'chaos-mesh', REPO / 'helm/chaos-mesh', '--namespace', 'chaos-mesh', '--create-namespace',
                  '-f', values, '--wait', '--timeout', '240s'], 250)
        host.kubectl('wait', '--for=condition=Ready', 'pod/target', 'pod/control', '-n', NAMESPACE, '--timeout=60s', seconds=65)
        pods = json.loads(host.kubectl('get', 'pods', '-A', '-o', 'json'))
        selected = select_pods(pods['items'])
        bpfki = pod_identity(selected['bpfki'], 'bpfki')
        binding = {key: bpfki[key] for key in ('pod', 'pod_uid', 'runtime_id', 'image', 'pod_image_id')}
        need(binding['image'] == bundle['images']['kernel']['tag'], 'gate sidecar is not the candidate image')
        for role in ('daemon', 'controller'):
            need(not selected[role]['spec'].get('hostNetwork'), 'chart bypassed Pod networking')
        webhooks = json.loads(host.kubectl('get', 'validatingwebhookconfigurations', 'chaos-mesh-validation', '-o', 'json'))
        need(all(w['failurePolicy'] == 'Fail' for w in webhooks['webhooks']), 'webhook does not fail closed')
        with (output / 'watchdog.log').open('wb') as log:
            supervisor = subprocess.Popen([sys.executable, str(Path(probe.__file__)), 'watchdog', profile, str(output)],
                                          stdin=subprocess.DEVNULL, stdout=log, stderr=log, start_new_session=True)
        ready = wait_file(output / 'watchdog-ready.json', time.monotonic() + 5, supervisor)
        need(ready['guest'] == asdict(guest), 'watchdog bound another guest')
        proof = json.loads(host.kubectl('exec', '-i', '-n', 'chaos-mesh', binding['pod'], '-c', 'bpfki', '--',
                                       'python3', '-B', '-', seconds=85, data=(inputs / 'qualify_bpf.py').read_bytes()))
        current = pod_identity(json.loads(host.kubectl('get', 'pod', binding['pod'], '-n', 'chaos-mesh', '-o', 'json')), 'bpfki')
        need(all(current[key] == value for key, value in binding.items()), 'sidecar changed during qualification')
        receipt = {'guest': asdict(guest), 'bpfki': binding, 'proof': proof}
        require_gate(receipt, guest, binding)
        host.install_json('gate', receipt)
        with (output / 'observer.log').open('wb') as log:
            process = subprocess.Popen(host.kube + ['exec', '-n', NAMESPACE, 'observer', '-c', 'observer', '--',
                                                   'python3', '-B', '/tools/guest.py'], stdout=log, stderr=log,
                                       stdin=subprocess.DEVNULL, start_new_session=True)
        end = time.monotonic() + 90
        while True:
            need(process.poll() is None and time.monotonic() < end, 'observer did not become ready')
            exists = host.ssh('test ! -f /var/lib/cm-minikube/evidence/ready.json || cat /var/lib/cm-minikube/evidence/ready.json', 10)
            if exists.strip():
                ready = json.loads(exists)
                break
            time.sleep(0.5)
        need(ready['guest'] == asdict(guest) and ready['bpfki'] == binding, 'observer readiness identity differs')
        write_json(output / 'arm.json', asdict(guest))
        lease = wait_file(output / 'armed.json', time.monotonic() + 5, supervisor)
        need(lease['guest'] == asdict(guest), 'watchdog active lease identity differs')
        guest_clock = float(host.ssh("cut -d ' ' -f 1 /proc/uptime"))
        remaining = lease['deadline_monotonic'] - time.monotonic()
        need(remaining > BUDGET.active - 15, 'active lease transport used too much time')
        host.install_json('active', {'guest': asdict(guest), 'deadline_boottime': guest_clock + remaining - 5})
        while process.poll() is None:
            need(supervisor.poll() is None, 'independent watchdog exited during the active lease')
            need(time.monotonic() < lease['deadline_monotonic'] - 2, 'observer exceeded its active lease')
            time.sleep(0.1)
        code = process.returncode
        result['recovery'] = json.loads(host.read_evidence('recovery.json'))
        (output / 'recovery.json').write_text(json.dumps(result['recovery'], indent=2))
        (output / 'lifecycle-events.jsonl').write_bytes(host.read_evidence('lifecycle-events.jsonl'))
        need(code == 0 and result['recovery']['ok'] is True, 'bounded lifecycle recovery was not proven')
    except BaseException as error:
        result['error'] = type(error).__name__ + ': ' + str(error)
    finally:
        try:
            probe.stop_domain(output, guest)
            result['containment']['ok'] = True
            (output / 'watchdog-done').touch()
        except BaseException as error:
            result['containment']['error'] = str(error)
        if process is not None and process.poll() is None:
            process.terminate()
            try:
                process.wait(timeout=2)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait(timeout=2)
        if supervisor is not None and result['containment']['ok']:
            try:
                supervisor.wait(timeout=2)
            except subprocess.TimeoutExpired:
                pass
        write_json(output / 'qualification.json', result)
    need(result['recovery']['ok'] and result['containment']['ok'], 'qualification failed; see separate recovery and containment results')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('profile')
    parser.add_argument('inputs', type=Path)
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    qualify(args.profile, args.inputs, args.output)
