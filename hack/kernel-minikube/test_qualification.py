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


import copy
from dataclasses import asdict
import errno
import hashlib
import http.client
import io
import json
from pathlib import Path
import socket
import tarfile
import tempfile
import threading
import time
import unittest
from unittest.mock import Mock, patch

import probe
import api
import cases
from identity import Guest, Workflow, archive_identity, verify_inputs


GUEST = Guest('cm-kvm-12-3', '11111111-1111-1111-1111-111111111111',
              '22222222-2222-2222-2222-222222222222', '33333333-3333-3333-3333-333333333333',
              '4' * 32, 123, '6.6.95', 'x86_64', '5' * 64)
BPFKI = {'pod': 'daemon', 'pod_uid': '6' * 36, 'runtime_id': 'containerd://' + '7' * 64}


def receipt():
    return {'guest': asdict(GUEST), 'bpfki': BPFKI,
            'proof': {'ok': True, 'helper_retained': True, 'detached': True,
                      'map_entries_before': 0, 'map_entries_after': 0, 'archive_uid': 0,
                      'header_release': '6.6.95', 'archive_sha256': '8' * 64,
                      'automatic_headers': True, 'entry': 'should_failslab'}}


class GateTests(unittest.TestCase):
    def setUp(self):
        self.clock = patch.object(cases, 'now', time.monotonic)
        self.clock.start()
        self.api_clock = patch.object(api, 'now', time.monotonic)
        self.api_clock.start()
        self.addCleanup(self.clock.stop)
        self.addCleanup(self.api_clock.stop)

    def client(self):
        client = api.APIClient.__new__(api.APIClient)
        client.armed = False
        client.request = Mock(return_value={})
        return client

    def test_failed_gate_cannot_create_kernelchaos(self):
        for field, wrong in [('ok', False), ('helper_retained', False), ('detached', False),
                             ('map_entries_before', 1), ('map_entries_after', 1), ('archive_uid', 10001),
                             ('header_release', 'other'), ('automatic_headers', False), ('entry', 'other')]:
            with self.subTest(field=field):
                client = self.client()
                proof = receipt()
                proof['proof'][field] = wrong
                with self.assertRaises((RuntimeError, ValueError)):
                    client.arm(proof, GUEST, BPFKI, {'guest': asdict(GUEST), 'deadline_boottime': time.monotonic() + 290})
                with self.assertRaises(RuntimeError):
                    client.create({}, time.monotonic() + 1)
                client.request.assert_not_called()

    def test_wrong_identity_or_missing_receipt_cannot_arm(self):
        for change in ('domain_uuid', 'boot_id'):
            proof = receipt()
            proof['guest'][change] = '0' * 36
            with self.assertRaises(ValueError):
                self.client().arm(proof, GUEST, BPFKI, {'guest': asdict(GUEST), 'deadline_boottime': time.monotonic() + 290})
        with self.assertRaises(KeyError):
            self.client().arm({}, GUEST, BPFKI, {})
        with self.assertRaises(ValueError):
            GUEST.check_live('wrong-boot', GUEST.machine_id, GUEST.initial_pid_namespace, '6.6.95', 'x86_64')
        with self.assertRaises(ValueError):
            GUEST.check_live(GUEST.boot_id, GUEST.machine_id, 456, '6.6.95', 'x86_64')

    def test_one_successful_lease_allows_one_unreplayed_mutation(self):
        client = self.client()
        lease = {'guest': asdict(GUEST), 'deadline_boottime': time.monotonic() + 290}
        client.arm(receipt(), GUEST, BPFKI, lease)
        with self.assertRaises(RuntimeError):
            client.arm(receipt(), GUEST, BPFKI, lease)
        client.request.side_effect = TimeoutError('ambiguous POST response')
        with self.assertRaises(TimeoutError):
            client.create({'kind': 'KernelChaos'}, time.monotonic() + 1)
        self.assertEqual(client.request.call_count, 1)

    def test_delete_carries_uid_precondition(self):
        client = self.client()
        client.delete('delete', GUEST.node_uid, time.monotonic() + 1)
        args = client.request.call_args.args
        self.assertEqual(args[0], 'DELETE')
        self.assertEqual(args[3]['preconditions'], {'uid': GUEST.node_uid})

    def test_operation_limit_prevents_another_worker_request(self):
        observer = cases.Observer({'deadline': time.monotonic() + 1}, Mock())
        observer.ask = Mock()
        observer.counts['target'] = 8
        with self.assertRaises(RuntimeError):
            observer.operation('target', False, time.monotonic() + 1)
        observer.ask.assert_not_called()
        observer.counts['target'] = 4
        observer.faults = 4
        with self.assertRaises(RuntimeError):
            observer.operation('target', True, time.monotonic() + 1)
        observer.ask.assert_not_called()

    def test_only_target_enomem_is_accepted_as_a_fault(self):
        observer = cases.Observer({'deadline': time.monotonic() + 1}, Mock())
        observer.ask = Mock(return_value={'seq': 1, 'close_errno': 0, 'started_ns': 0,
                                          'finished_ns': 100, 'result': -1, 'errno': errno.EIO})
        with self.assertRaises(RuntimeError):
            observer.operation('target', True, time.monotonic() + 1)
        self.assertEqual(observer.ask.call_count, 1)


class HTTPTests(unittest.TestCase):
    def setUp(self):
        self.clock = patch.object(cases, 'now', time.monotonic)
        self.clock.start()
        self.addCleanup(self.clock.stop)

    def connection(self, chunks):
        client, server = socket.socketpair()
        connection = http.client.HTTPConnection('local', timeout=0.2)
        connection.sock = client
        connection.connect = lambda: None
        def serve():
            try:
                server.recv(16384)
                for delay, value in chunks:
                    time.sleep(delay)
                    server.sendall(value)
            except OSError:
                pass
            finally:
                server.close()
        worker = threading.Thread(target=serve, daemon=True)
        worker.start()
        self.addCleanup(connection.close)
        self.addCleanup(lambda: worker.join(timeout=1))
        return connection

    def test_absolute_deadline_stops_slow_headers(self):
        chunks = [(0, b'HTTP/1.1 200 OK\r\nX-Slow: ')] + [(0.02, b'a')] * 30
        connection = self.connection(chunks)
        started = time.monotonic()
        with patch.object(api, 'now', time.monotonic), self.assertRaises(Exception):
            response = api.Response(connection, 'GET', '/', None, {}, started + 0.12)
            try:
                response.json()
            finally:
                response.close()
        self.assertLess(time.monotonic() - started, 0.6)

    def test_absolute_deadline_stops_slow_partial_json(self):
        chunks = [(0, b'HTTP/1.1 200 OK\r\nContent-Length: 40\r\n\r\n')] + [(0.02, b' ')] * 30
        connection = self.connection(chunks)
        started = time.monotonic()
        with patch.object(api, 'now', time.monotonic):
            response = api.Response(connection, 'GET', '/', None, {}, started + 0.12)
            try:
                with self.assertRaises(Exception):
                    response.json()
            finally:
                response.close()
        self.assertLess(time.monotonic() - started, 0.6)

    def test_missing_requires_a_real_notfound_status(self):
        client = api.APIClient.__new__(api.APIClient)
        response = Mock()
        response.response.status = 404
        client.open = Mock(return_value=response)
        for value in ({'kind': 'Status', 'reason': 'NotFound', 'code': 404},
                      {'kind': 'Status', 'reason': 'Forbidden', 'code': 403}, {'message': 'NotFound'}):
            response.json.return_value = value
            if value.get('reason') == 'NotFound':
                self.assertIsNone(client.get('delete', time.monotonic() + 1, missing=True))
            else:
                with self.assertRaises(RuntimeError):
                    client.get('delete', time.monotonic() + 1, missing=True)
        response.response.status = 403
        with self.assertRaises(api.APIError):
            client.get('delete', time.monotonic() + 1, missing=True)


class ArtifactTests(unittest.TestCase):
    def create_inputs(self, root):
        workflow = Workflow('12', '3', 'a' * 40, 'b' * 40)
        layer = b'fixed layer content'
        config = json.dumps({'architecture': 'amd64', 'os': 'linux',
                             'rootfs': {'diff_ids': ['sha256:' + hashlib.sha256(layer).hexdigest()]}}).encode()
        images = {}
        for role in ('kernel', 'controller', 'daemon', 'worker'):
            path = root / (role + '.tar.gz')
            with tarfile.open(path, 'w:gz') as archive:
                for name, data in [('manifest.json', json.dumps([{'Config': 'config.json', 'Layers': ['layer.tar']}]).encode()),
                                   ('config.json', config), ('layer.tar', layer)]:
                    item = tarfile.TarInfo(name)
                    item.size = len(data)
                    archive.addfile(item, io.BytesIO(data))
            images[role] = dict(archive_identity(path), archive=path.name, tag='test:' + role)
        value = {'workflow': asdict(workflow), 'images': images, 'files': {}}
        path = root / 'inputs.json'
        path.write_text(json.dumps(value))
        return workflow, value, path

    def test_wrong_run_source_archive_config_and_diffids_are_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            workflow, value, path = self.create_inputs(Path(directory))
            self.assertEqual(verify_inputs(path, workflow), value)
            for key in ('run_id', 'attempt', 'checkout', 'head'):
                changed = copy.deepcopy(value)
                changed['workflow'][key] = 'wrong'
                path.write_text(json.dumps(changed))
                with self.assertRaises(ValueError):
                    verify_inputs(path, workflow)
            for key in ('archive_sha256', 'config_digest', 'diff_ids'):
                changed = copy.deepcopy(value)
                changed['images']['kernel'][key] = 'wrong'
                path.write_text(json.dumps(changed))
                with self.assertRaises(ValueError):
                    verify_inputs(path, workflow)


class ContainmentTests(unittest.TestCase):
    def test_stop_targets_bound_uuid_before_any_slow_profile_delete(self):
        calls = []
        def command(output, name, args, timeout=30):
            calls.append(args)
            value = GUEST.domain_uuid if 'domuuid' in args else 'shut off'
            return {'returncode': 0, 'stdout': value, 'stderr': ''}
        with patch.object(probe, 'command', side_effect=command):
            probe.stop_domain(Path('/unused'), GUEST)
            probe.stop_domain(Path('/unused'), GUEST)
        self.assertEqual(calls[0][-2:], ['destroy', GUEST.domain_uuid])
        self.assertTrue(all(GUEST.profile not in args for args in calls if 'destroy' in args))
        self.assertFalse(any('delete' in args for args in calls))

    def test_reused_profile_is_not_deleted(self):
        def command(output, name, args, timeout=30):
            value = 'another-domain' if 'domuuid' in args else 'shut off'
            return {'returncode': 0, 'stdout': value, 'stderr': ''}
        with patch.object(probe, 'command', side_effect=command), self.assertRaisesRegex(ValueError, 'UUID changed'):
            probe.stop_domain(Path('/unused'), GUEST)

    def test_watchdog_stops_vm_at_fixed_lease_before_delete(self):
        clock = [0.0]
        calls = []
        def command(output, name, args, timeout=30):
            calls.append((clock[0], args))
            value = GUEST.domain_uuid if 'domuuid' in args else 'shut off'
            return {'returncode': 0, 'stdout': value, 'stderr': ''}
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory)
            (output / 'guest.json').write_text(json.dumps(asdict(GUEST)))
            (output / 'arm.json').write_text(json.dumps(asdict(GUEST)))
            with patch.object(probe.time, 'monotonic', side_effect=lambda: clock[0]), \
                    patch.object(probe.time, 'sleep', side_effect=lambda _: clock.__setitem__(0, clock[0] + 1)), \
                    patch.object(probe, 'command', side_effect=command):
                probe.watchdog(GUEST.profile, output)
            self.assertEqual(calls[0], (300, ['virsh', '-c', 'qemu:///system', 'destroy', GUEST.domain_uuid]))
            self.assertEqual(calls[-1][1], ['minikube', '-p', GUEST.profile, 'delete'])
            self.assertEqual(json.loads((output / 'armed.json').read_text())['deadline_monotonic'], 300)


if __name__ == '__main__':
    unittest.main()
