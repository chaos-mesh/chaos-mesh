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
import ctypes
from dataclasses import asdict
import errno
import hashlib
import http.client
import io
import json
from pathlib import Path
import socket
import struct
import tarfile
import tempfile
import threading
import time
import unittest
from unittest.mock import MagicMock, Mock, patch

import probe
import qualify_bpf
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


class LoadedProgramTests(unittest.TestCase):
    SYMBOLS = ('ffffffff81002000 T __bpf_call_base\n'
               'ffffffff81001000 t bpf_override_return\n')

    def setUp(self):
        self.raw = struct.pack('<BBhi', 0x85, 0, 0, -4096)
        self.calls = []
        self.kallsyms = Mock()
        self.kallsyms.read_text.return_value = self.SYMBOLS
        self.files = patch.object(qualify_bpf, 'Path', return_value=self.kallsyms)
        self.library = patch.object(qualify_bpf.ctypes, 'CDLL', return_value=Mock(syscall=self.syscall))
        self.files.start()
        self.library.start()
        self.addCleanup(self.files.stop)
        self.addCleanup(self.library.stop)

    def syscall(self, number, command, pointer, size):
        self.assertEqual((number, command, size), (321, 15, ctypes.sizeof(qualify_bpf.InfoAttr)))
        attr = ctypes.cast(pointer, ctypes.POINTER(qualify_bpf.InfoAttr)).contents
        self.assertEqual(attr.info_len, ctypes.sizeof(qualify_bpf.ProgramInfo))
        info = qualify_bpf.ProgramInfo.from_address(attr.info)
        self.calls.append((info.xlated_prog_len, info.jited_prog_len))
        if info.xlated_prog_len:
            ctypes.memmove(info.xlated_prog_insns, self.raw, min(info.xlated_prog_len, len(self.raw)))
        if info.jited_prog_len and not info.jited_prog_insns:
            ctypes.set_errno(errno.EFAULT)
            return -1
        info.id = 123
        info.xlated_prog_len = len(self.raw)
        info.jited_prog_len = 64
        return 0

    def test_jited_program_requests_only_translated_bytes_and_records_live_helper_identity(self):
        for immediate, symbols in ((-4096, self.SYMBOLS),
                                   (4096, self.SYMBOLS.replace('81002000', '81000000'))):
            with self.subTest(immediate=immediate):
                self.raw = struct.pack('<BBhi', 0x85, 0, 0, immediate)
                self.calls.clear()
                self.kallsyms.read_text.return_value = symbols
                proof = qualify_bpf.loaded_program(7)
                self.assertEqual(proof, {
                    'program_id': 123, 'helper_retained': True,
                    'translated_sha256': hashlib.sha256(self.raw).hexdigest(),
                    'helper_identity': {'source': '/proc/kallsyms', 'symbols': {
                        '__bpf_call_base': '0xffffffff81002000' if immediate < 0 else '0xffffffff81000000',
                        'bpf_override_return': '0xffffffff81001000'}, 'call_immediate': immediate}})
                self.assertEqual(self.calls, [(0, 0), (len(self.raw), 0)])

    def test_original_helper_id_wrong_offset_and_pseudo_calls_are_rejected(self):
        for instruction in ((0x85, 0, 0, 58), (0x85, 0, 0, 1), (0x85, 0, 0, 0),
                            (0x85, 0, 0, -8192), (0x85, 0x10, 0, -4096),
                            (0x85, 0x20, 0, -4096), (0xB7, 0, 0, -4096), (0x85, 0, 1, -4096)):
            with self.subTest(instruction=instruction):
                self.raw = struct.pack('<BBhi', *instruction)
                with self.assertRaisesRegex(RuntimeError, 'override helper was removed'):
                    qualify_bpf.loaded_program(7)

    def test_unreadable_missing_zero_ambiguous_and_invalid_symbols_are_rejected(self):
        self.raw += struct.pack('<BBhi', 0x85, 0, 0, 58)
        invalid = [PermissionError('kallsyms denied'), '',
                   self.SYMBOLS.replace('ffffffff81001000', '0000000000000000'),
                   self.SYMBOLS.replace('ffffffff81002000', '0000000000000000'),
                   self.SYMBOLS.replace('ffffffff81001000', 'invalid'),
                   self.SYMBOLS + 'ffffffff81003000 t bpf_override_return\n',
                   self.SYMBOLS + 'ffffffff81003000 T __bpf_call_base\n',
                   self.SYMBOLS.replace('81001000', '81002000'),
                   self.SYMBOLS.replace('ffffffff81001000', 'ffffffff01001000'),
                   self.SYMBOLS.replace('ffffffff81002000', 'ffffffff00002000'),
                   self.SYMBOLS.replace('ffffffff81001000', '10000000000000000'),
                   self.SYMBOLS.replace('t bpf_override_return', 'd bpf_override_return'),
                   self.SYMBOLS.replace('bpf_override_return\n', 'bpf_override_return [module]\n'),
                   self.SYMBOLS.replace('bpf_override_return\n', '__bpf_override_return\n')]
        for symbols in invalid:
            with self.subTest(symbols=symbols):
                self.kallsyms.read_text.side_effect = symbols if isinstance(symbols, OSError) else None
                self.kallsyms.read_text.return_value = symbols
                with self.assertRaises((RuntimeError, OSError, ValueError)):
                    qualify_bpf.loaded_program(7)


class QualifierCleanupTests(unittest.TestCase):
    def run_qualifier(self, leaked_fd=None, retained_probe=False):
        release = '#define UTS_RELEASE "6.6.95"'
        config = '\n'.join('CONFIG_' + name + '=y' for name in (
            'BPF', 'BPF_SYSCALL', 'KPROBES', 'BPF_EVENTS', 'KPROBE_EVENTS',
            'BPF_KPROBE_OVERRIDE', 'FUNCTION_ERROR_INJECTION', 'IKHEADERS')).encode()
        files = MagicMock()
        files.stat.return_value.st_uid = 0
        files.read_bytes.return_value = qualify_bpf.gzip.compress(config)
        files.exists.return_value = False
        files.is_dir.return_value = True
        files.__truediv__.return_value.read_text.return_value = release
        headers = MagicMock()
        members = [tarfile.TarInfo(name) for name in (
            'include/linux/kconfig.h', 'include/generated/autoconf.h', 'include/generated/utsrelease.h')]
        headers.__iter__.return_value = iter(members)
        headers.extractfile.return_value.read.return_value = release.encode()
        descriptors = []
        bpf = MagicMock()
        bpf.kprobe_fds = {}
        bpf.num_open_kprobes.side_effect = lambda: len(bpf.kprobe_fds)
        bpf.__getitem__.return_value.items.return_value = []

        def create_probe(**kwargs):
            for _ in range(4):
                descriptor = open('/dev/null', 'rb')
                self.addCleanup(descriptor.close)
                descriptors.append(descriptor)
            bpf.load_func.return_value.fd = descriptors[0].fileno()
            bpf.__getitem__.return_value.map_fd = descriptors[1].fileno()
            return bpf

        def attach(**kwargs):
            bpf.kprobe_fds[b'p_should_failslab'] = {b'qualify': descriptors[2].fileno()}

        def cleanup():
            if not retained_probe:
                bpf.kprobe_fds[b'p_should_failslab'].clear()
            for index, descriptor in enumerate(descriptors):
                if index != leaked_fd:
                    descriptor.close()

        bpf.attach_kprobe.side_effect = attach
        bpf.cleanup.side_effect = cleanup
        with patch.dict(qualify_bpf.os.environ, {}, clear=True), \
                patch.object(qualify_bpf.os, 'geteuid', return_value=0), \
                patch.object(qualify_bpf.os, 'uname', return_value=Mock(release='6.6.95', machine='x86_64')), \
                patch.object(qualify_bpf.signal, 'signal'), patch.object(qualify_bpf.signal, 'alarm'), \
                patch.object(qualify_bpf, 'guest_allowlist', return_value='should_failslab ERRNO'), \
                patch.object(qualify_bpf, 'prepare_tracefs'), \
                patch.object(qualify_bpf, 'Path', return_value=files), \
                patch.object(qualify_bpf.tarfile, 'open') as archive, \
                patch.object(qualify_bpf, 'loaded_program', return_value={'helper_retained': True}), \
                patch.object(qualify_bpf.os, 'listdir', side_effect=lambda _: [str(fd.fileno()) for fd in descriptors]), \
                patch.object(qualify_bpf.os, 'readlink', return_value='anon_inode:bpf'), \
                patch.dict('sys.modules', {'bcc': Mock(BPF=Mock(side_effect=create_probe))}), \
                patch('sys.argv', ['qualify_bpf.py', GUEST.boot_id, str(GUEST.initial_pid_namespace)]), \
                patch('sys.stdout', new_callable=io.StringIO) as output:
            archive.return_value.__enter__.return_value = headers
            status = qualify_bpf.main()
        bpf.attach_kprobe.assert_called_once_with(event='should_failslab', fn_name='qualify')
        bpf.cleanup.assert_called_once_with()
        result = json.loads(output.getvalue())
        self.assertNotIn('error', result)
        self.assertEqual(len(result['owned_fds']), 4)
        return status, result

    def test_empty_bcc_event_key_after_cleanup_allows_success(self):
        status, result = self.run_qualifier()
        self.assertEqual(status, 0, result)
        self.assertTrue(result['ok'])
        self.assertTrue(result['detached'])
        self.assertNotIn('cleanup_error', result)

    def test_each_owned_fd_must_be_closed_even_when_probe_entries_are_empty(self):
        for index in range(4):
            with self.subTest(descriptor=index):
                status, result = self.run_qualifier(leaked_fd=index)
                self.assertEqual(status, 1, result)
                self.assertFalse(result['ok'])
                self.assertFalse(result['detached'])
                self.assertIn('qualifier FD remains open', result['cleanup_error'])

    def test_retained_probe_entry_fails_even_when_all_owned_fds_are_closed(self):
        status, result = self.run_qualifier(retained_probe=True)
        self.assertEqual(status, 1, result)
        self.assertFalse(result['ok'])
        self.assertFalse(result['detached'])
        self.assertIn('entry probe remained attached', result['cleanup_error'])


class AllowlistTests(unittest.TestCase):
    def setUp(self):
        self.paths = {}
        for name in ('/proc/self/ns/pid', '/proc/1/ns/pid'):
            self.paths[name] = Mock()
            self.paths[name].stat.return_value.st_ino = GUEST.initial_pid_namespace
        self.paths['/proc/sys/kernel/random/boot_id'] = Mock()
        self.paths['/proc/sys/kernel/random/boot_id'].read_text.return_value = GUEST.boot_id + '\n'
        self.paths['/proc/1/mountinfo'] = Mock()
        self.paths['/proc/1/mountinfo'].read_text.return_value = (
            '27 19 0:8 / /sys/kernel/debug rw,nosuid,nodev - debugfs debugfs rw\n')
        self.allowlist = '/proc/1/root/sys/kernel/debug/error_injection/list'
        self.paths[self.allowlist] = Mock()
        self.paths[self.allowlist].read_text.return_value = 'should_failslab ERRNO\n'
        self.files = patch.object(qualify_bpf, 'Path', side_effect=self.paths.__getitem__)
        self.files.start()
        self.addCleanup(self.files.stop)

    def read(self):
        return qualify_bpf.guest_allowlist(GUEST.boot_id, GUEST.initial_pid_namespace)

    def test_reads_guest_debugfs_without_container_debugfs(self):
        self.assertEqual(self.read(), 'should_failslab ERRNO\n')
        self.paths[self.allowlist].read_text.assert_called_once_with()

    def test_wrong_boot_or_pid_namespace_rejects_before_allowlist(self):
        for name in ('/proc/self/ns/pid', '/proc/1/ns/pid'):
            with self.subTest(path=name):
                self.paths[name].stat.return_value.st_ino += 1
                with self.assertRaisesRegex(RuntimeError, 'PID namespace'):
                    self.read()
                self.paths[name].stat.return_value.st_ino -= 1
        self.paths['/proc/sys/kernel/random/boot_id'].read_text.return_value = 'other-boot'
        with self.assertRaisesRegex(RuntimeError, 'boot ID'):
            self.read()
        self.paths[self.allowlist].read_text.assert_not_called()

    def test_wrong_mount_point_type_or_root_rejects_before_allowlist(self):
        for line in (
                '27 19 0:8 / /sys/kernel/debug-other rw - debugfs debugfs rw',
                '27 19 0:8 / /sys/kernel/debug rw - tmpfs debugfs rw',
                '27 19 0:8 /subtree /sys/kernel/debug rw - debugfs debugfs rw',
                '27 19 0:8 / /sys/kernel/tracing rw - tracefs tracefs rw',
                ''):
            with self.subTest(mountinfo=line):
                self.paths['/proc/1/mountinfo'].read_text.return_value = line
                with self.assertRaisesRegex(RuntimeError, 'debugfs'):
                    self.read()
        self.paths[self.allowlist].read_text.assert_not_called()

    def test_invalid_guest_proof_never_imports_bcc(self):
        import builtins
        for name, value, message in (('/proc/sys/kernel/random/boot_id', 'other-boot', 'boot ID'),
                                     ('/proc/1/mountinfo', '', 'debugfs')):
            with self.subTest(path=name):
                original = self.paths[name].read_text.return_value
                self.paths[name].read_text.return_value = value
                with patch.dict(qualify_bpf.os.environ, {}, clear=True), \
                        patch.object(qualify_bpf.os, 'geteuid', return_value=0), \
                        patch.object(qualify_bpf.os, 'uname', return_value=Mock(release='6.6.95', machine='x86_64')), \
                        patch.object(qualify_bpf.signal, 'signal'), patch.object(qualify_bpf.signal, 'alarm'), \
                        patch('sys.argv', ['qualify_bpf.py', GUEST.boot_id, str(GUEST.initial_pid_namespace)]), \
                        patch('sys.stdout', new_callable=io.StringIO) as output, \
                        patch('builtins.__import__', wraps=builtins.__import__) as imported:
                    self.assertEqual(qualify_bpf.main(), 1)
                    result = json.loads(output.getvalue())
                    self.assertFalse(result['ok'])
                    self.assertIn(message, result['error'])
                    self.assertFalse(any(call.args[0] == 'bcc' for call in imported.call_args_list))
                self.paths[name].read_text.return_value = original
        self.paths[self.allowlist].read_text.assert_not_called()


class TracefsTests(unittest.TestCase):
    SYSFS = '30 20 0:5 / /sys rw,nosuid,nodev,noexec - sysfs sysfs rw\n'
    TRACEFS = '31 30 0:9 / /sys/kernel/tracing rw,nosuid,nodev,noexec - tracefs tracefs rw\n'

    def setUp(self):
        self.paths = {name: Mock() for name in (
            '/proc/self/ns/mnt', '/proc/1/ns/mnt', '/proc/self/mountinfo', '/proc/1/mountinfo',
            '/sys/kernel/debug/tracing', '/sys/kernel/tracing', '/sys/kernel/tracing/kprobe_events')}
        self.paths['/proc/self/ns/mnt'].stat.return_value.st_ino = 22
        self.paths['/proc/1/ns/mnt'].stat.return_value.st_ino = 11
        self.paths['/proc/self/mountinfo'].read_text.return_value = self.SYSFS
        self.paths['/proc/1/mountinfo'].read_text.return_value = self.TRACEFS
        self.paths['/sys/kernel/debug/tracing'].exists.return_value = False
        self.paths['/sys/kernel/tracing'].is_dir.return_value = True
        self.paths['/sys/kernel/tracing'].resolve.return_value = '/sys/kernel/tracing'
        self.paths['/sys/kernel/tracing/kprobe_events'].open.side_effect = lambda mode: io.BytesIO(b'')
        self.libc = Mock()
        def mount(*args):
            self.paths['/proc/self/mountinfo'].read_text.return_value += self.TRACEFS
            return 0
        self.libc.mount.side_effect = mount
        self.files = patch.object(qualify_bpf, 'Path', side_effect=self.paths.__getitem__)
        self.library = patch.object(qualify_bpf.ctypes, 'CDLL', return_value=self.libc)
        self.files.start()
        self.library.start()
        self.addCleanup(self.files.stop)
        self.addCleanup(self.library.stop)

    def prepare(self):
        self.proof = {}
        return qualify_bpf.prepare_tracefs(self.proof)

    def test_mounts_only_private_container_tracefs_with_fixed_flags(self):
        proof = self.prepare()
        self.libc.mount.assert_called_once_with(b'tracefs', b'/sys/kernel/tracing', b'tracefs', 14, None)
        self.assertTrue(proof['mounted'])
        self.assertEqual(proof['after'][0]['filesystem'], 'tracefs')
        self.assertEqual(proof['namespaces'], {'self': 22, '1': 11})
        self.paths['/sys/kernel/tracing/kprobe_events'].open.assert_called_once_with('rb')

    def test_existing_usable_tracefs_is_not_mounted_again(self):
        self.paths['/proc/self/mountinfo'].read_text.return_value += self.TRACEFS
        proof = self.prepare()
        self.assertFalse(proof['mounted'])
        self.libc.mount.assert_not_called()

    def test_guest_mount_namespace_never_mounts(self):
        self.paths['/proc/self/ns/mnt'].stat.return_value.st_ino = 11
        with self.assertRaisesRegex(RuntimeError, 'guest mount namespace'):
            self.prepare()
        self.libc.mount.assert_not_called()

    def test_shared_wrong_parent_and_descendant_mounts_are_rejected(self):
        for mounts in (self.SYSFS.replace(' - ', ' shared:8 - '),
                       self.SYSFS.replace(' - sysfs ', ' - tmpfs '),
                       self.SYSFS + self.TRACEFS.replace('/sys/kernel/tracing', '/sys/kernel/tracing/nested'),
                       self.SYSFS + self.TRACEFS.replace('/sys/kernel/tracing', '/sys/kernel')):
            with self.subTest(mounts=mounts):
                self.paths['/proc/self/mountinfo'].read_text.return_value = mounts
                with self.assertRaises(RuntimeError):
                    self.prepare()
        self.libc.mount.assert_not_called()

    def test_guest_superblock_nondefault_options_are_rejected(self):
        for options in ('ro', 'rw,sync', 'rw,mand', 'rw,lazytime'):
            with self.subTest(options=options):
                self.paths['/proc/1/mountinfo'].read_text.return_value = self.TRACEFS.replace('tracefs rw', 'tracefs ' + options)
                with self.assertRaisesRegex(RuntimeError, 'only rw superblock options'):
                    self.prepare()
        self.libc.mount.assert_not_called()

    def test_symlink_target_and_shadowed_bcc_path_are_rejected(self):
        self.paths['/sys/kernel/tracing'].resolve.return_value = '/somewhere/else'
        with self.assertRaisesRegex(RuntimeError, 'symlink'):
            self.prepare()
        self.paths['/sys/kernel/debug/tracing'].exists.return_value = True
        with self.assertRaisesRegex(RuntimeError, 'shadows'):
            self.prepare()
        self.libc.mount.assert_not_called()

    def test_mount_failure_and_incomplete_verification_fail_the_gate(self):
        for status, message in ((-1, 'cannot mount'), (0, 'verification failed')):
            with self.subTest(status=status):
                self.libc.mount.side_effect = None
                self.libc.mount.return_value = status
                with self.assertRaisesRegex(RuntimeError, message):
                    self.prepare()
        self.paths['/sys/kernel/tracing/kprobe_events'].open.assert_not_called()


    def test_mount_failure_stops_before_bcc_and_preserves_prerequisite_evidence(self):
        import builtins
        self.libc.mount.side_effect = None
        self.libc.mount.return_value = -1
        with patch.dict(qualify_bpf.os.environ, {}, clear=True), \
                patch.object(qualify_bpf.os, 'geteuid', return_value=0), \
                patch.object(qualify_bpf.os, 'uname', return_value=Mock(release='6.6.95', machine='x86_64')), \
                patch.object(qualify_bpf.signal, 'signal'), patch.object(qualify_bpf.signal, 'alarm'), \
                patch.object(qualify_bpf, 'guest_allowlist', return_value='should_failslab ERRNO'), \
                patch('sys.argv', ['qualify_bpf.py', GUEST.boot_id, str(GUEST.initial_pid_namespace)]), \
                patch('sys.stdout', new_callable=io.StringIO) as output, \
                patch('builtins.__import__', wraps=builtins.__import__) as imported:
            self.assertEqual(qualify_bpf.main(), 1)
            result = json.loads(output.getvalue())
            self.assertIn('cannot mount', result['error'])
            self.assertFalse(result['ok'])
            self.assertFalse(result['tracefs']['mounted'])
            self.assertEqual(result['tracefs']['guest'][0]['superblock_options'], ['rw'])
            self.assertFalse(any(call.args[0] == 'bcc' for call in imported.call_args_list))


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
