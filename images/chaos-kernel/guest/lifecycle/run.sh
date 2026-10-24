#!/usr/bin/env bash
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

set -euo pipefail
if [[ $# != 2 || ${GITHUB_ACTIONS:-} != true || ${RUNNER_ENVIRONMENT:-} != github-hosted ||
      ${RUNNER_OS:-} != Linux || ${RUNNER_ARCH:-} != X64 ||
      ${RUNNER_TEMP:-} != /home/runner/work/_temp || $EUID == 0 ||
      $(uname -s) != Linux || $(uname -m) != x86_64 ]]; then
  echo 'Usage on a disposable GitHub-hosted amd64 runner: run.sh EXTRACTED_BUNDLE OUTPUT_DIR' >&2
  exit 2
fi
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
exec python3 - "$script_dir" "$@" <<'PYTHON'
import base64
import hashlib
import io
import json
import os
from pathlib import Path, PurePosixPath
import re
import selectors
import shutil
import signal
import stat
import subprocess
import sys
import tarfile
import time

script, bundle, output = map(lambda p: Path(p).resolve(), sys.argv[1:])
GiB = 1024 ** 3
limit = 8 * 1024 ** 2
process = None
output.mkdir(parents=True, exist_ok=True)


def need(ok, message):
    if not ok:
        raise RuntimeError(message)


def sha(path):
    digest = hashlib.sha256()
    with path.open('rb') as source:
        for block in iter(lambda: source.read(1024 * 1024), b''):
            digest.update(block)
    return digest.hexdigest()


def write(name, value):
    (output / name).write_text(json.dumps(value, sort_keys=True, indent=2) + '\n')


def interrupted(number, frame):
    raise RuntimeError('launcher interrupted by signal ' + str(number))


signal.signal(signal.SIGTERM, interrupted)
signal.signal(signal.SIGINT, interrupted)
summary = {'passed': False, 'qemu_exit_code': None, 'guest_exit_code': None, 'error': '', 'launched': False}
try:
    need(bundle.is_relative_to(Path(os.environ['RUNNER_TEMP'])) and output.is_relative_to(Path(os.environ['RUNNER_TEMP'])),
         'bundle and output must be owned runner temporary paths')
    need(not output.is_relative_to(bundle) and not bundle.is_relative_to(output), 'output overlaps sealed bundle')
    manifest = json.loads((bundle / 'bundle.json').read_text())
    need(manifest['schema_version'] == 1, 'unsupported bundle')
    envelope = {'cpus': 2, 'memory_mib': 6144, 'disk_bytes': 16 * GiB, 'guest_deadline_seconds': 840,
                'outer_deadline_seconds': 900, 'termination_grace_seconds': 20, 'nic': 'none',
                'serial_and_evidence_limit_bytes': limit, 'read_only_shares': ['headers', 'payload']}
    need(manifest['envelope'] == envelope, 'launch envelope changed')
    actual = {}
    for path in sorted(bundle.rglob('*')):
        relative = path.relative_to(bundle).as_posix()
        if relative == 'bundle.json':
            continue
        mode = path.lstat().st_mode
        record = {'mode': stat.S_IMODE(mode)}
        if stat.S_ISLNK(mode):
            target = os.readlink(path)
            need(not target.startswith('/') and path.resolve().is_relative_to(bundle) and path.exists(), 'unsafe bundle link')
            if relative.startswith('headers/') or relative.startswith('payload/'):
                share = bundle / relative.split('/', 1)[0]
                need(path.resolve().is_relative_to(share), 'link escapes its finite share')
            record.update(kind='symlink', target=target)
        elif stat.S_ISDIR(mode):
            record['kind'] = 'directory'
        else:
            need(stat.S_ISREG(mode), 'special bundle entry')
            record.update(kind='file', sha256=sha(path), size=path.stat().st_size)
        actual[relative] = record
    need(actual == manifest['files'], 'bundle inventory, modes or hashes changed')
    need(set(path.name for path in bundle.iterdir()) == {'bundle.json', 'rootfs.ext4', 'bzImage', 'headers', 'payload', 'review'},
         'unexpected launch input')
    need(sha(bundle / 'payload/node.json') == manifest['node_sha256'], 'node identity changed')
    need(sha(bundle / 'payload/lifecycle/run.sh') == sha(script / 'run.sh'), 'launcher differs from prepared source')
    checkout = subprocess.check_output(['git', '-C', str(script.parents[3]), 'rev-parse', 'HEAD'], text=True).strip()
    need(checkout == manifest['source_sha'], 'launcher checkout differs from prepared merge')
    disk = bundle / 'rootfs.ext4'
    need(disk.stat().st_size == 16 * GiB and sha(disk) == manifest['capacity']['pristine_disk_sha256'], 'pristine disk changed')
    need(manifest['capacity']['projected_guest_bytes'] <= 16 * GiB, 'prepared guest does not fit')
    need(all(',' not in str(path) and '\n' not in str(path) for path in [bundle, output]), 'unsafe QEMU path')
    for path in Path('/proc').glob('[0-9]*/cmdline'):
        try:
            parts = path.read_bytes().split(b'\0')
        except (FileNotFoundError, PermissionError, ProcessLookupError):
            continue
        executable = os.path.basename(os.fsdecode(parts[0]))
        busy = executable in {'make', 'gmake', 'ninja', 'cc1', 'cc1plus', 'gcc', 'g++', 'clang', 'clang++', 'qemu-system-x86_64'}
        busy = busy or (executable == 'docker' and any(part in {b'build', b'buildx'} for part in parts[1:3]))
        need(not busy, 'concurrent build or guest detected')
    memory = int(re.search(r'^MemAvailable:\s+(\d+) kB$', Path('/proc/meminfo').read_text(), re.M)[1]) * 1024
    allocated = disk.stat().st_blocks * 512
    remaining = max(0, 16 * GiB - allocated)
    artifact_growth = limit + 1024 * 1024
    free = shutil.disk_usage(bundle).free
    capacity = {'mem_available_bytes': memory, 'minimum_mem_available_bytes': 8 * GiB,
                'host_free_bytes': free, 'remaining_disk_allocation_bytes': remaining,
                'host_reserve_bytes': 4 * GiB, 'artifact_growth_bytes': artifact_growth,
                'disk_allocated_bytes': allocated}
    write('launch-capacity.json', capacity)
    need(memory >= 8 * GiB, 'host MemAvailable is below 8 GiB')
    need(free >= remaining + 4 * GiB + artifact_growth, 'host lacks remaining allocation, reserve and evidence space')
    command = ['qemu-system-x86_64', '-machine', 'q35,accel=tcg', '-cpu', 'max', '-m', '6144', '-smp', '2',
        '-nodefaults', '-no-reboot', '-display', 'none', '-monitor', 'none', '-serial', 'stdio', '-nic', 'none',
        '-kernel', str(bundle / 'bzImage'),
        '-append', 'console=ttyS0 root=/dev/vda rw init=/opt/cm-lifecycle/init.sh panic=1 oops=panic systemd.unified_cgroup_hierarchy=1 cgroup_no_v1=all cm_lifecycle=' + manifest['node_sha256'],
        '-drive', 'file=' + str(disk) + ',format=raw,if=virtio',
        '-virtfs', 'local,path=' + str(bundle / 'headers') + ',mount_tag=cm-headers,security_model=none,readonly=on',
        '-virtfs', 'local,path=' + str(bundle / 'payload') + ',mount_tag=cm-payload,security_model=none,readonly=on']
    with (output / 'run-started.json').open('x') as record:
        record.write(json.dumps({'node_sha256': manifest['node_sha256'], 'command': command, 'acceleration': 'tcg'}) + '\n')
    started = time.monotonic()
    process = subprocess.Popen(command, stdin=subprocess.DEVNULL, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, start_new_session=True)
    summary['launched'] = True
    write('qemu-process.json', {'pid': process.pid, 'started_monotonic': started})
    selector = selectors.DefaultSelector()
    selector.register(process.stdout, selectors.EVENT_READ)
    serial_bytes = 0
    with (output / 'serial.log').open('wb') as destination:
        while selector.get_map():
            remaining_time = 900 - (time.monotonic() - started)
            need(remaining_time > 0, 'outer 900-second deadline expired')
            for key, _ in selector.select(min(1, remaining_time)):
                chunk = os.read(key.fd, 65536)
                if not chunk:
                    selector.unregister(key.fileobj)
                    continue
                need(serial_bytes + len(chunk) <= limit, 'serial evidence exceeds 8 MiB; incomplete evidence fails')
                destination.write(chunk)
                destination.flush()
                serial_bytes += len(chunk)
    summary['qemu_exit_code'] = process.wait(timeout=max(0.01, 900 - (time.monotonic() - started)))
    serial = (output / 'serial.log').read_bytes().replace(b'\r', b'')
    matches = list(re.finditer(rb'(?:^|\n)CM_LIFECYCLE_EVIDENCE_BEGIN ([0-9a-f]{64}) ([0-9]+)\n', serial))
    need(len(matches) == 1 and serial.count(b'CM_LIFECYCLE_EVIDENCE_END\n') == 1, 'missing or repeated evidence framing')
    match = matches[0]
    encoded, trailer = serial[match.end():].split(b'CM_LIFECYCLE_EVIDENCE_END\n', 1)
    raw = base64.b64decode(b''.join(encoded.splitlines()), validate=True)
    need(len(raw) == int(match[2]) and hashlib.sha256(raw).hexdigest().encode() == match[1], 'serial archive length or hash mismatch')
    extracted = output / 'guest'
    extracted.mkdir()
    names = set()
    evidence_bytes = 0
    with tarfile.open(fileobj=io.BytesIO(raw), mode='r:gz') as archive:
        for entry in archive:
            path = PurePosixPath(entry.name)
            need(entry.isfile() and not path.is_absolute() and '..' not in path.parts and len(path.parts) == 1,
                 'unexpected guest evidence entry')
            need(entry.name not in names, 'duplicate guest evidence')
            names.add(entry.name)
            evidence_bytes += entry.size
            need(evidence_bytes + serial_bytes <= limit, 'serial plus required evidence exceeds 8 MiB')
            with archive.extractfile(entry) as source, (extracted / entry.name).open('xb') as destination:
                shutil.copyfileobj(source, destination)
    need(summary['qemu_exit_code'] == 0, 'QEMU exited unsuccessfully')
    exits = re.findall(rb'(?:^|\n)CM_LIFECYCLE_EXIT=([01])(?:\n|$)', trailer)
    need(len(exits) == 1, 'missing or repeated actual guest exit')
    summary['guest_exit_code'] = int(exits[0])
    need(summary['guest_exit_code'] == 0 and b'reboot: Power down' in trailer, 'guest did not complete successful poweroff')
    need(b'CM_LIFECYCLE_WATCHDOG_EXPIRED' not in serial and b'CM_LIFECYCLE_EVIDENCE_INVALID' not in serial, 'guest watchdog or evidence failed')
    required = {'guest-result.json', 'service-exit.json', 'lifecycle-events.jsonl', 'boot.json',
                'imported-images.json', 'addon-kube-proxy.json', 'addon-coredns.json', 'node-ready.json',
                'admission.json', 'chart-review.json', 'chart-ready.json', 'fixture-ready.json',
                'final-module-state.json', 'cleanup-module-state.json'}
    need(required <= names, 'missing required guest evidence files')
    result = json.loads((extracted / 'guest-result.json').read_text())
    service = json.loads((extracted / 'service-exit.json').read_text())
    need(service['ok'] is True and service['exit_code'] == 'exited' and service['exit_status'] == '0'
         and service['boot_seconds'] < 840, 'actual service exit or boot deadline failed')
    need(result['ok'] is True and result['cleanup']['ok'] is True and not result['error'], 'case or cleanup failed')
    need(len(result['cases']) == 2 and [case['kind'] for case in result['cases']] == ['expiration', 'delete'] and
         all(case['ok'] is True for case in result['cases']), 'both lifecycle cases did not pass')
    node = json.loads((bundle / 'payload/node.json').read_text())
    for key in ['source_sha', 'bpfki_commit', 'kernel', 'worker']:
        need(result[key] == node[key], 'guest result source binding differs: ' + key)
    raw_events = (extracted / 'lifecycle-events.jsonl').read_text()
    need(raw_events.endswith('\n'), 'partial lifecycle event record')
    events = [json.loads(line) for line in raw_events.splitlines()]
    need([event['event'] for event in events] == list(range(1, len(events) + 1)), 'lifecycle event sequence has gaps')
    outcomes = [event['value'] for event in events if event['kind'] == 'lifecycle_result']
    need(len(outcomes) == 1 and outcomes[0]['ok'] is True and outcomes[0]['cases'] == result['cases'], 'raw case verdict differs')
    need(events[-1]['kind'] == 'cleanup' and events[-1]['value'] == result['cleanup'], 'raw cleanup result differs')
    for case in result['cases']:
        samples = [event['value']['response'] for event in events
                   if event['kind'] == 'native_sample' and event['case_uid'] == case['uid']]
        need(samples and all(sample['ok'] is True for sample in samples), 'missing or failed raw native sample')
        for role in ['target', 'control']:
            need(sum(sample['command'] == 'operate ' + role for sample in samples) == 4, 'missing required worker operations')
        for command in ['map absent', 'map active', 'map recovered']:
            need(any(sample['command'] == command for sample in samples), 'missing required map observation')
    summary.update(passed=True, serial_bytes=serial_bytes, evidence_bytes=evidence_bytes,
                   elapsed_seconds=time.monotonic() - started, node_sha256=manifest['node_sha256'],
                   cleanup=result['cleanup'], cases=result['cases'], service_exit=service)
except BaseException as error:
    summary['error'] = str(error)
    raise
finally:
    if process is not None and process.poll() is None:
        try:
            os.killpg(process.pid, signal.SIGTERM)
        except ProcessLookupError:
            pass
        try:
            process.wait(timeout=20)
        except subprocess.TimeoutExpired:
            try:
                os.killpg(process.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
            process.wait(timeout=5)
    if process is not None:
        summary['qemu_exit_code'] = process.returncode
    write('result.json', summary)
PYTHON
