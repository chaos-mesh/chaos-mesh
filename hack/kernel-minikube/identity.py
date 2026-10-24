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
import re
import tarfile
import tempfile
from dataclasses import dataclass


def need(ok, reason):
    if not ok:
        raise ValueError(reason)


def write_json(path, value):
    path = Path(path)
    with tempfile.NamedTemporaryFile(mode='w', dir=path.parent, prefix='.' + path.name, delete=False) as stream:
        temporary = Path(stream.name)
        try:
            stream.write(json.dumps(value, sort_keys=True, indent=2) + '\n')
            stream.close()
            temporary.replace(path)
        finally:
            temporary.unlink(missing_ok=True)


def sha(path):
    result = hashlib.sha256()
    with Path(path).open('rb') as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b''):
            result.update(block)
    return result.hexdigest()


@dataclass(frozen=True)
class Workflow:
    run_id: str
    attempt: str
    checkout: str
    head: str

    @classmethod
    def current(cls):
        value = cls(os.environ['GITHUB_RUN_ID'], os.environ['GITHUB_RUN_ATTEMPT'],
                    os.environ['GITHUB_SHA'], os.environ.get('PR_HEAD_SHA') or os.environ['GITHUB_SHA'])
        need(all(re.fullmatch(r'[0-9]+', item) for item in (value.run_id, value.attempt)), 'invalid run identity')
        need(all(re.fullmatch(r'[0-9a-f]{40}', item) for item in (value.checkout, value.head)), 'invalid source identity')
        return value

    @property
    def profile(self):
        return f'cm-kvm-{self.run_id}-{self.attempt}'


def archive_identity(path):
    with tarfile.open(path) as archive:
        manifest = json.load(archive.extractfile('manifest.json'))
        need(len(manifest) == 1, 'image archive must contain exactly one image')
        entry = manifest[0]
        data = archive.extractfile(entry['Config']).read(4 * 1024 * 1024 + 1)
        need(len(data) <= 4 * 1024 * 1024, 'image config exceeds bound')
        config = json.loads(data)
        need(config['architecture'] == 'amd64' and config['os'] == 'linux', 'image is not linux/amd64')
        diffs = config['rootfs']['diff_ids']
        need(len(diffs) == len(entry['Layers']), 'image layer count mismatch')
        for name, expected in zip(entry['Layers'], diffs):
            digest = hashlib.sha256()
            with archive.extractfile(name) as stream:
                for block in iter(lambda: stream.read(1024 * 1024), b''):
                    digest.update(block)
            need('sha256:' + digest.hexdigest() == expected, 'archive layer content differs from diffID')
    return {'config_digest': 'sha256:' + hashlib.sha256(data).hexdigest(),
            'diff_ids': diffs, 'archive_sha256': sha(path)}


def verify_inputs(path, workflow):
    path = Path(path)
    value = json.loads(path.read_text())
    need(value['workflow'] == workflow.__dict__, 'producer is from another run, attempt or source')
    need(set(value['images']) == {'kernel', 'controller', 'daemon', 'worker'}, 'incomplete candidate image set')
    for role, record in value['images'].items():
        need(record['archive'] == role + '.tar.gz', 'invalid artifact path')
        actual = archive_identity(path.parent / record['archive'])
        need(all(record[key] == item for key, item in actual.items()), 'candidate image content changed')
    for name, expected in value['files'].items():
        need(Path(name).name == name and name not in {'.', '..'}, 'invalid tool artifact path')
        need(sha(path.parent / name) == expected, 'tool or verifier digest changed: ' + name)
    return value


@dataclass(frozen=True)
class Guest:
    profile: str
    domain_uuid: str
    boot_id: str
    node_uid: str
    machine_id: str
    initial_pid_namespace: int
    release: str
    architecture: str
    nonce: str

    @classmethod
    def read(cls, value, profile):
        guest = cls(**value)
        need(guest.profile == profile and re.fullmatch(r'cm-kvm-[0-9]+-[0-9]+', profile), 'wrong guest profile')
        for item in (guest.domain_uuid, guest.boot_id, guest.node_uid):
            need(re.fullmatch(r'[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}', item), 'invalid guest UUID')
        need(re.fullmatch(r'[0-9a-f]{32}', guest.machine_id), 'invalid guest machine ID')
        need(guest.release == '6.6.95' and guest.architecture == 'x86_64', 'unexpected stock kernel')
        need(type(guest.initial_pid_namespace) is int and guest.initial_pid_namespace > 0, 'invalid initial PID namespace')
        need(re.fullmatch(r'[0-9a-f]{64}', guest.nonce), 'invalid guest nonce')
        return guest

    def check_live(self, boot_id, machine_id, namespace, release, architecture):
        need((boot_id, machine_id, namespace, release, architecture) ==
             (self.boot_id, self.machine_id, self.initial_pid_namespace, self.release, self.architecture),
             'live guest differs from host-bound identity')


def require_gate(receipt, guest, bpfki):
    need(receipt['guest'] == guest.__dict__ and receipt['bpfki'] == bpfki, 'qualification belongs to another guest or container')
    proof = receipt['proof']
    need(proof['ok'] is True and proof['helper_retained'] is True and proof['detached'] is True,
         'override-capable qualification or detach was not proven')
    need(proof['map_entries_before'] == proof['map_entries_after'] == 0, 'qualifier map was populated')
    need(proof['archive_uid'] == 0 and proof['header_release'] == guest.release and
         re.fullmatch(r'[0-9a-f]{64}', proof['archive_sha256']), 'header provenance was not proven')
    need(proof['automatic_headers'] is True and proof['entry'] == 'should_failslab', 'wrong header or attachment path')
