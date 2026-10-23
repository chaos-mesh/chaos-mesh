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


"""Bind the one workflow-owned stock VM and enforce exact-domain containment."""
import argparse
import json
import os
from pathlib import Path
import re
import shutil
import shlex
import signal
import subprocess
import sys
import time
import uuid

sys.path.insert(0, str(Path(__file__).resolve().parents[2] / 'images/chaos-kernel/guest/lifecycle'))
from checks import BUDGET
from identity import Guest, write_json

KVM_API = "import fcntl, os; f = os.open('/dev/kvm', os.O_RDWR); print(fcntl.ioctl(f, 0xAE00, 0)); os.close(f)"
OPTIONS = ('BPF', 'BPF_SYSCALL', 'KPROBES', 'BPF_KPROBE_OVERRIDE',
           'FUNCTION_ERROR_INJECTION', 'BPF_EVENTS', 'KPROBE_EVENTS', 'IKHEADERS', 'DEBUG_FS',
           'FAULT_INJECTION', 'FAILSLAB')


def command(output, name, args, timeout=30):
    result = {'argv': args}
    try:
        with subprocess.Popen(args, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                              text=True, start_new_session=True) as proc:
            try:
                stdout, stderr = proc.communicate(timeout=timeout)
            except subprocess.TimeoutExpired:
                os.killpg(proc.pid, signal.SIGKILL)
                stdout, stderr = proc.communicate()
                result['timeout'] = True
            result.update(returncode=proc.returncode, stdout=stdout, stderr=stderr)
    except OSError as error:
        result.update(returncode=None, stdout='', stderr=str(error))
    (output / (name + '.json')).write_text(json.dumps(result, indent=2) + '\n')
    return result


def need(ok, reason):
    if not ok:
        raise ValueError(reason)


def record_for(profile, output):
    run_id, attempt = os.environ['GITHUB_RUN_ID'], os.environ['GITHUB_RUN_ATTEMPT']
    need(bool(re.fullmatch(r'[0-9]+', run_id) and re.fullmatch(r'[0-9]+', attempt)), 'invalid run identity')
    need(profile == f'cm-kvm-{run_id}-{attempt}', 'profile differs from this run attempt')
    path = output / 'result.json'
    if path.exists():
        record = json.loads(path.read_text())
        need(record['profile'] == profile and record['run_id'] == run_id and
             record['run_attempt'] == attempt, 'evidence belongs to another attempt')
        return record
    return dict(schema_version=1, profile=profile, run_id=run_id, run_attempt=attempt,
                github_sha=os.environ.get('GITHUB_SHA'), head_sha=os.environ.get('PR_HEAD_SHA'),
                repository=os.environ.get('GITHUB_REPOSITORY'), fault_qualification='untested',
                host={'status': 'untested'}, transport={'status': 'untested', 'image': 'stock'},
                fault_prerequisites={'status': 'untested'}, cleanup={'status': 'untested'})


def host(profile, output, record):
    need(sys.platform == 'linux' and os.environ.get('RUNNER_ENVIRONMENT') == 'github-hosted',
         'host probe requires a GitHub-hosted Linux runner')
    cpu = Path('/proc/cpuinfo').read_text()
    memory = Path('/proc/meminfo').read_text()
    available = int(re.search(r'^MemAvailable:\s+(\d+)', memory, re.M)[1])
    facts = dict(cpu_virtualization=bool(re.search(r'\b(vmx|svm)\b', cpu)),
                 cpus=len(os.sched_getaffinity(0)), available_memory_kib=available,
                 free_disk_bytes=shutil.disk_usage(Path.home()).free)
    record['host'].update(facts)
    need(facts['cpu_virtualization'], 'CPU does not expose vmx or svm')
    need(facts['cpus'] >= 2 and available >= 8 * 1024**2 and
         facts['free_disk_bytes'] >= 20 * 1024**3, 'need 2 CPUs, 8 GiB available RAM, and 20 GiB free disk')
    need(Path('/dev/kvm').is_char_device(), '/dev/kvm is not a character device')
    kvm = command(output, 'host-kvm-api', ['sudo', '-n', 'python3', '-c', KVM_API])
    need(kvm['returncode'] == 0 and kvm['stdout'].strip() == '12', 'KVM API unavailable even with sudo')
    record['host'].update(status='proven', kvm_api=12, kvm_access='sudo; runner ACL configured next')
    checkout = command(output, 'checkout', ['git', 'rev-parse', 'HEAD'])
    need(checkout['returncode'] == 0, 'cannot record checkout')
    record['checkout_sha'] = checkout['stdout'].strip()
    need(not any((Path.home() / '.minikube' / kind / profile).exists()
                 for kind in ('profiles', 'machines')), 'owned profile already exists before startup')


def node_identity(raw, profile):
    nodes = json.loads(raw)['items']
    need(len(nodes) == 1 and nodes[0]['metadata']['name'] == profile, 'unexpected node identity')
    node = nodes[0]
    need(any(c['type'] == 'Ready' and c['status'] == 'True' for c in node['status']['conditions']), 'node is not Ready')
    info = node['status']['nodeInfo']
    need(info['kubeletVersion'] == 'v1.34.5' and info['containerRuntimeVersion'].startswith('containerd://'),
         'unexpected Kubernetes version or runtime')
    return info


def prerequisites(raw):
    values = dict(line.split('=', 1) for line in raw.splitlines())
    expected = {'config_' + option for option in OPTIONS} | {'symbol', 'allowlist', 'headers'}
    need(expected <= values.keys(), 'incomplete guest evidence')
    checks = {key: values[key] for key in expected}
    need(set(checks.values()) <= {'present', 'absent', 'unavailable'}, 'invalid guest evidence state')
    required = [value for key, value in checks.items() if key not in {'headers', 'config_FAULT_INJECTION', 'config_FAILSLAB'}]
    status = 'unsupported' if 'absent' in required else ('unknown' if 'unavailable' in required else 'observed')
    return dict(status=status, checks=checks, kernel=values['kernel'], config_source=values['config_source'])


def guest(profile, output, record):
    need(record['host']['status'] == 'proven', 'host capability was not proven')
    nodes = command(output, 'nodes', ['minikube', '-p', profile, 'kubectl', '--',
                                    'get', 'nodes', '-o', 'json', '--request-timeout=20s'], 45)
    need(nodes['returncode'] == 0, 'cannot observe Kubernetes node')
    info = node_identity(nodes['stdout'], profile)
    script = r'''set -eu
k=$(uname -r)
printf 'kernel=%s\n' "$k"
if test -r /proc/config.gz; then
  config=$(zcat /proc/config.gz); source=/proc/config.gz
elif test -r "/boot/config-$k"; then
  config=$(cat "/boot/config-$k"); source="/boot/config-$k"
else config=''; source=unavailable; fi
printf 'config_source=%s\n' "$source"
for option in OPTIONS; do
  state=unavailable
  if test "$source" != unavailable; then
    state=absent
    if printf '%s\n' "$config" | grep -qx "CONFIG_${option}=y"; then state=present; fi
  fi
  printf 'config_%s=%s\n' "$option" "$state"
done
for item in symbol allowlist; do
  if test "$item" = symbol; then path=/proc/kallsyms; else path=/sys/kernel/debug/error_injection/list; fi
  state=unavailable
  if test -r "$path"; then
    state=absent
    if test "$item" = symbol; then
      if grep -qw should_failslab "$path"; then state=present; fi
    elif grep -Eq '(^|[[:space:]])should_failslab[[:space:]].*ERRNO' "$path"; then state=present; fi
  fi
  printf '%s=%s\n' "$item" "$state"
done
headers=absent
build="/lib/modules/$k/build"
if test -r "$build/include/generated/autoconf.h" && test -r "$build/include/config/kernel.release"; then
  if test "$(cat "$build/include/config/kernel.release")" = "$k"; then headers=present; fi
fi
printf 'headers=%s\n' "$headers"
'''.replace('OPTIONS', ' '.join(OPTIONS))
    observed = command(output, 'guest-prerequisites', ['minikube', '-p', profile, 'ssh', '--', 'sudo sh -c ' + shlex.quote(script)], 45)
    need(observed['returncode'] == 0, 'guest observation failed')
    record['fault_prerequisites'] = prerequisites(observed['stdout'])
    namespaces = command(output, 'guest-namespaces', ['minikube', '-p', profile, 'ssh', '--',
                         'sudo readlink /proc/1/ns/pid /proc/1/ns/mnt /proc/self/ns/pid /proc/self/ns/mnt'])
    need(namespaces['returncode'] == 0, 'namespace observation failed')
    record['transport'].update(status='proven', node=profile, node_info=info)
    domain = command(output, 'domain-uuid', ['virsh', '-c', 'qemu:///system', 'domuuid', profile], 10)
    need(domain['returncode'] == 0, 'cannot bind owned domain UUID')
    identity = command(output, 'guest-identity', ['minikube', '-p', profile, 'ssh', '--',
                       "sudo sh -c 'cat /proc/sys/kernel/random/boot_id /etc/machine-id; stat -Lc %i /proc/1/ns/pid; uname -r; uname -m; cat /sys/class/dmi/id/product_uuid'"], 20)
    need(identity['returncode'] == 0, 'cannot bind guest identity')
    boot, machine, namespace, release, architecture, dmi_uuid = identity['stdout'].split()
    need(dmi_uuid.lower() == domain['stdout'].strip(), 'domain and live DMI UUID differ')
    node = json.loads(nodes['stdout'])['items'][0]
    guest = Guest.read(dict(profile=profile, domain_uuid=domain['stdout'].strip(), boot_id=boot,
                            node_uid=node['metadata']['uid'], machine_id=machine,
                            initial_pid_namespace=int(namespace), release=release, architecture=architecture,
                            nonce=uuid.uuid4().hex + uuid.uuid4().hex), profile)
    need(info['bootID'] == boot and info['machineID'] == machine, 'node and guest identity differ')
    write_json(output / 'guest.json', guest.__dict__)
    record['guest'] = guest.__dict__


def cleanup(profile, output, record):
    outcome = os.environ.get('START_OUTCOME', 'skipped')
    record['workflow_steps'] = {key: os.environ.get(key, 'unknown') for key in ('HOST_OUTCOME', 'SETUP_OUTCOME', 'START_OUTCOME')}
    if outcome != 'skipped' and record['transport']['status'] == 'untested':
        record['transport'].update(status='failed', reason='startup or guest observation did not complete')
    paths = [Path.home() / '.minikube' / kind / profile for kind in ('profiles', 'machines')]
    if outcome == 'skipped':
        need(not any(path.exists() for path in paths), 'profile exists although startup was skipped')
        record['cleanup'] = dict(status='not_created', reason='start action skipped', domain_check='untested')
        return
    guest_path = output / 'guest.json'
    if guest_path.exists():
        guest = Guest.read(json.loads(guest_path.read_text()), profile)
        stop_domain(output, guest)
    else:
        domain = command(output, 'cleanup-unbound-domain', ['virsh', '-c', 'qemu:///system', 'domuuid', profile], 10)
        if domain['returncode'] == 0:
            need(bool(re.fullmatch(r'[0-9a-f-]{36}', domain['stdout'].strip())), 'invalid startup domain UUID')
            command(output, 'stop-startup-domain', ['virsh', '-c', 'qemu:///system', 'destroy', domain['stdout'].strip()], 10)
    deleted = (command(output, 'minikube-delete', ['minikube', '-p', profile, 'delete'], 90)
               if any(path.exists() for path in paths) or not guest_path.exists() else {'returncode': 0})
    domains = command(output, 'remaining-domains', ['virsh', '-c', 'qemu:///system', 'list', '--all', '--name'], 15)
    networks = command(output, 'remaining-networks', ['virsh', '-c', 'qemu:///system', 'net-list', '--all', '--name'], 15)
    baseline = output / 'networks-before.txt'
    new_networks = sorted(set(networks['stdout'].split()) - set(baseline.read_text().split())) if baseline.exists() else None
    networks_absent = networks['returncode'] == 0 and new_networks == []
    absent = not any(path.exists() for path in paths)
    domain_absent = domains['returncode'] == 0 and profile not in domains['stdout'].splitlines()
    qualification = output / 'qualification.json'
    if qualification.exists():
        record['fault_qualification'] = 'proven' if json.loads(qualification.read_text())['recovery']['ok'] else 'failed'
    record['cleanup'] = dict(status='removed' if absent and domain_absent and networks_absent and deleted['returncode'] == 0 else 'failed',
                             profile_absent=absent, domain_absent=domain_absent, new_networks_remaining=new_networks,
                             delete_returncode=deleted['returncode'])
    need(record['cleanup']['status'] == 'removed', 'owned profile cleanup did not complete')


def stop_domain(output, guest):
    command(output, 'stop-domain', ['virsh', '-c', 'qemu:///system', 'destroy', guest.domain_uuid], 5)
    state = command(output, 'stopped-domain-state', ['virsh', '-c', 'qemu:///system', 'domstate', guest.domain_uuid], 5)
    if state['returncode'] == 0:
        need(state['stdout'].strip() == 'shut off', 'exact domain did not stop')
    else:
        domains = command(output, 'stop-domain-list', ['virsh', '-c', 'qemu:///system', 'list', '--all', '--uuid'], 5)
        need(domains['returncode'] == 0 and guest.domain_uuid not in domains['stdout'].split(), 'cannot establish domain absence')
    domain = command(output, 'stop-domain-identity', ['virsh', '-c', 'qemu:///system', 'domuuid', guest.profile], 5)
    need(domain['returncode'] != 0 or domain['stdout'].strip() == guest.domain_uuid, 'owned profile domain UUID changed')


def watchdog(profile, output):
    guest = Guest.read(json.loads((output / 'guest.json').read_text()), profile)
    deadline = time.monotonic() + 180
    write_json(output / 'watchdog-ready.json', {'guest': guest.__dict__, 'startup_deadline': deadline})
    active = False
    while time.monotonic() < deadline:
        if (output / 'watchdog-done').exists():
            return
        request = output / 'arm.json'
        if not active and request.exists():
            need(json.loads(request.read_text()) == guest.__dict__, 'watchdog lease identity changed')
            deadline = time.monotonic() + BUDGET.active
            active = True
            write_json(output / 'armed.json', {'guest': guest.__dict__, 'deadline_monotonic': deadline})
        time.sleep(0.05)
    commands = output / 'watchdog-commands'
    commands.mkdir(exist_ok=True)
    outcome = {'expired': True, 'active': active, 'domain_uuid': guest.domain_uuid}
    try:
        stop_domain(commands, guest)
        outcome['stopped'] = True
        deleted = command(commands, 'watchdog-delete', ['minikube', '-p', profile, 'delete'], 90)
        outcome['delete_returncode'] = deleted['returncode']
    except BaseException as error:
        outcome['error'] = str(error)
    write_json(output / 'watchdog.json', outcome)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('stage', choices=('host', 'guest', 'cleanup', 'watchdog'))
    parser.add_argument('profile')
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    record = record_for(args.profile, args.output)
    args.output.mkdir(parents=True, exist_ok=True)
    if args.stage == 'watchdog':
        watchdog(args.profile, args.output)
        return 0
    stage = {'host': 'host', 'guest': 'fault_prerequisites', 'cleanup': 'cleanup'}[args.stage]
    try:
        globals()[args.stage](args.profile, args.output, record)
    except (ValueError, OSError, KeyError, TypeError) as error:
        record[stage].update(status='failed', reason=str(error))
        print(str(error), file=sys.stderr)
        return 1
    finally:
        (args.output / 'result.json').write_text(json.dumps(record, indent=2, sort_keys=True) + '\n')
        if args.stage == 'cleanup' and os.environ.get('GITHUB_STEP_SUMMARY'):
            with open(os.environ['GITHUB_STEP_SUMMARY'], 'a') as summary:
                summary.write('Stock minikube qualification. Recovery and VM cleanup are reported separately.\n\n')
                for key in ('host', 'transport', 'fault_prerequisites', 'cleanup'):
                    summary.write(f"- {key}: {record[key]['status']}\n")
    return 0


if __name__ == '__main__':
    sys.exit(main())
