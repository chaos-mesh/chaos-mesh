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


import ctypes
import errno
import gzip
import hashlib
import json
import os
from pathlib import Path
import re
import signal
import struct
import sys
import tarfile


def need(ok, message):
    if not ok:
        raise RuntimeError(message)


class ProgramInfo(ctypes.Structure):
    _fields_ = [('type', ctypes.c_uint32), ('id', ctypes.c_uint32), ('tag', ctypes.c_ubyte * 8),
                ('jited_prog_len', ctypes.c_uint32), ('xlated_prog_len', ctypes.c_uint32),
                ('jited_prog_insns', ctypes.c_uint64), ('xlated_prog_insns', ctypes.c_uint64)]


class InfoAttr(ctypes.Structure):
    _fields_ = [('fd', ctypes.c_uint32), ('info_len', ctypes.c_uint32), ('info', ctypes.c_uint64)]


def loaded_program(fd):
    libc = ctypes.CDLL(None, use_errno=True)
    info = ProgramInfo()
    attr = InfoAttr(fd, ctypes.sizeof(info), ctypes.addressof(info))
    need(libc.syscall(321, 15, ctypes.byref(attr), ctypes.sizeof(attr)) == 0, 'cannot inspect loaded BPF program')
    need(0 < info.xlated_prog_len <= 65536, 'unexpected translated program length')
    instructions = ctypes.create_string_buffer(info.xlated_prog_len)
    info.xlated_prog_insns = ctypes.addressof(instructions)
    need(libc.syscall(321, 15, ctypes.byref(attr), ctypes.sizeof(attr)) == 0, 'cannot read loaded BPF instructions')
    raw = instructions.raw[:info.xlated_prog_len]
    retained = any(code == 0x85 and registers == 0 and immediate == 58
                   for code, registers, offset, immediate in struct.iter_unpack('<BBhi', raw))
    need(retained, 'override helper was removed from loaded bytecode')
    return {'program_id': info.id, 'helper_retained': retained,
            'translated_sha256': hashlib.sha256(raw).hexdigest()}


def mountinfo(pid):
    mounts = []
    for line in Path('/proc/' + pid + '/mountinfo').read_text().splitlines():
        fields, filesystem = (part.split() for part in line.split(' - ', 1))
        paths = [re.sub(r'\\([0-7]{3})', lambda match: chr(int(match[1], 8)), value)
                 for value in fields[3:5]]
        mounts.append({'id': fields[0], 'root': paths[0], 'path': paths[1],
                       'options': fields[5].split(','), 'propagation': fields[6:],
                       'filesystem': filesystem[0], 'superblock_options': filesystem[2].split(',')})
    return mounts


def guest_allowlist(boot_id, initial_pid_namespace):
    need(Path('/proc/sys/kernel/random/boot_id').read_text().strip() == boot_id,
         'qualifier boot ID differs from bound guest')
    need(all(Path(path).stat().st_ino == initial_pid_namespace
             for path in ('/proc/self/ns/pid', '/proc/1/ns/pid')),
         'qualifier is not in the bound guest initial PID namespace')
    debugfs = [mount for mount in mountinfo('1') if mount['path'] == '/sys/kernel/debug']
    need(len(debugfs) == 1 and debugfs[0]['root'] == '/' and debugfs[0]['filesystem'] == 'debugfs',
         'guest PID 1 does not expose the debugfs root at /sys/kernel/debug')
    return Path('/proc/1/root/sys/kernel/debug/error_injection/list').read_text()


def prepare_tracefs(result):
    target = '/sys/kernel/tracing'
    selected = '/sys/kernel/debug/tracing' if Path('/sys/kernel/debug/tracing').exists() else target
    namespaces = {pid: Path('/proc/' + pid + '/ns/mnt').stat().st_ino for pid in ('self', '1')}
    need(namespaces['self'] != namespaces['1'], 'tracefs preparation is in the guest mount namespace')
    before = mountinfo('self')
    existing = [mount for mount in before if mount['path'] == selected]
    result.update(path=selected, namespaces=namespaces, mounted=False, before=existing)
    if existing:
        need(len(existing) == 1 and existing[0]['root'] == '/' and existing[0]['filesystem'] == 'tracefs',
             'BCC-selected tracing path is not a tracefs root')
        with Path(selected + '/kprobe_events').open('rb') as events:
            events.read(1)
        result['after'] = existing
        return result
    need(selected == target, 'unusable debugfs tracing path shadows /sys/kernel/tracing')
    need(Path(target).is_dir() and str(Path(target).resolve()) == target, 'tracefs target traverses a symlink or is absent')
    need(not any(mount['path'] == target or mount['path'].startswith(target + '/') for mount in before),
         'tracefs target already contains a mount')
    ancestors = [mount for mount in before if target.startswith(mount['path'].rstrip('/') + '/')]
    parent = max(ancestors, key=lambda mount: len(mount['path']))
    result['parent'] = parent
    need(parent['path'] == '/sys' and parent['root'] == '/' and parent['filesystem'] == 'sysfs' and
         not any(option.startswith('shared:') for option in parent['propagation']),
         'tracefs target is not on container sysfs without shared propagation')
    guest = [mount for mount in mountinfo('1') if mount['filesystem'] == 'tracefs' and mount['root'] == '/']
    result['guest'] = guest
    need(guest and all(mount['superblock_options'] == ['rw'] for mount in guest),
         'guest tracefs must expose only rw superblock options')
    libc = ctypes.CDLL(None, use_errno=True)
    libc.mount.argtypes = [ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p, ctypes.c_ulong, ctypes.c_void_p]
    libc.mount.restype = ctypes.c_int
    need(libc.mount(b'tracefs', target.encode(), b'tracefs', 14, None) == 0,
         'cannot mount container tracefs: ' + os.strerror(ctypes.get_errno()))
    result['mounted'] = True
    after = [mount for mount in mountinfo('self') if mount['path'] == target]
    result['after'] = after
    need(len(after) == 1 and after[0]['root'] == '/' and after[0]['filesystem'] == 'tracefs' and
         {'rw', 'nosuid', 'nodev', 'noexec'} <= set(after[0]['options']) and
         not any(option.startswith('shared:') for option in after[0]['propagation']),
         'container tracefs mount verification failed')
    need(not Path('/sys/kernel/debug/tracing').exists(), 'BCC tracing path changed during preparation')
    with Path(target + '/kprobe_events').open('rb') as events:
        events.read(1)
    return result


def main():
    result = {'ok': False, 'detached': False, 'entry': 'should_failslab'}
    probe = None
    owned_fds = []
    try:
        signal.signal(signal.SIGALRM, lambda *_: (_ for _ in ()).throw(TimeoutError('qualifier deadline')))
        signal.alarm(70)
        need(os.geteuid() == 0 and os.uname().release == '6.6.95' and os.uname().machine == 'x86_64',
             'qualifier is not in the stock guest')
        need(not any(key in os.environ for key in ('BCC_KERNEL_SOURCE', 'BCC_KERNEL_MODULES_SUFFIX', 'TMPDIR')),
             'BCC header lookup was overridden')
        allowlist = guest_allowlist(sys.argv[1], int(sys.argv[2]))
        result['tracefs'] = {}
        prepare_tracefs(result['tracefs'])
        config = gzip.decompress(Path('/proc/config.gz').read_bytes())
        options = dict(line.decode().split('=', 1) for line in config.splitlines() if line.startswith(b'CONFIG_') and b'=' in line)
        for option in ('BPF', 'BPF_SYSCALL', 'KPROBES', 'BPF_EVENTS', 'KPROBE_EVENTS',
                       'BPF_KPROBE_OVERRIDE', 'FUNCTION_ERROR_INJECTION', 'IKHEADERS'):
            need(options.get('CONFIG_' + option) == 'y', 'missing config ' + option)
        result['config_sha256'] = hashlib.sha256(config).hexdigest()
        result['optional_config'] = {name: options.get('CONFIG_' + name, 'absent')
                                     for name in ('FAILSLAB', 'FAULT_INJECTION')}
        need(any('should_failslab' in line.split() and 'ERRNO' in line for line in allowlist.splitlines()),
             'should_failslab entry is not ERRNO allowlisted')
        archive = Path('/sys/kernel/kheaders.tar.xz')
        result['archive_uid'] = archive.stat().st_uid
        need(result['archive_uid'] == 0, 'embedded headers are not root owned')
        result['archive_sha256'] = hashlib.sha256(archive.read_bytes()).hexdigest()
        with tarfile.open(archive, 'r:xz') as headers:
            members = {item.name.removeprefix('./'): item for item in headers}
            need(len(members) < 100000, 'header archive member bound exceeded')
            for name in ('include/linux/kconfig.h', 'include/generated/autoconf.h', 'include/generated/utsrelease.h'):
                need(name in members, 'missing embedded header ' + name)
            release = headers.extractfile(members['include/generated/utsrelease.h']).read(1024).decode().strip()
        need(release == '#define UTS_RELEASE "6.6.95"', 'embedded header release differs')
        result['header_release'] = '6.6.95'
        extracted = Path('/tmp/kheaders-6.6.95')
        need(not extracted.exists(), 'automatic header extraction cannot be distinguished from a preexisting cache')
        for path in ('source', 'build'):
            need(not Path('/lib/modules/6.6.95', path, 'include/linux/kconfig.h').exists(),
                 'conventional headers would bypass embedded archive qualification')
        from bcc import BPF
        before = set()
        for fd in os.listdir('/proc/self/fd'):
            try:
                os.readlink('/proc/self/fd/' + fd)
                before.add(fd)
            except FileNotFoundError:
                pass
        probe = BPF(text='''#include <uapi/linux/ptrace.h>
BPF_HASH(never_enabled, u32, u32, 1);
int qualify(struct pt_regs *ctx) {
    u32 key = 0;
    u32 *enabled = never_enabled.lookup(&key);
    if (enabled && *enabled) return bpf_override_return(ctx, -12);
    return 0;
}''')
        need(extracted.is_dir() and extracted.stat().st_uid == 0, 'BCC did not automatically extract embedded headers')
        need((extracted / 'include/generated/utsrelease.h').read_text().strip() == release,
             'BCC selected headers differ from archive release')
        result.update(automatic_headers=True, extracted_path=str(extracted))
        table = probe['never_enabled']
        result['map_entries_before'] = len(list(table.items()))
        need(result['map_entries_before'] == 0, 'qualifier map is not empty')
        function = probe.load_func('qualify', BPF.KPROBE)
        result.update(loaded_program(function.fd))
        probe.attach_kprobe(event='should_failslab', fn_name='qualify')
        need(probe.num_open_kprobes() == 1, 'entry probe was not attached')
        result['map_entries_after'] = len(list(table.items()))
        need(result['map_entries_after'] == 0, 'qualifier map changed')
        for fd in set(os.listdir('/proc/self/fd')) - before:
            try:
                target = os.readlink('/proc/self/fd/' + fd)
                if target.startswith('anon_inode:'):
                    owned_fds.append(int(fd))
            except FileNotFoundError:
                pass
        need(function.fd in owned_fds and table.map_fd in owned_fds, 'owned program or map FD was not captured')
        result['owned_fds'] = owned_fds
    except BaseException as error:
        result['error'] = type(error).__name__ + ': ' + str(error)
    finally:
        if probe is not None:
            try:
                probe.cleanup()
                need(probe.num_open_kprobes() == 0, 'entry probe remained attached')
                for fd in owned_fds:
                    try:
                        os.fstat(fd)
                    except OSError as error:
                        need(error.errno == errno.EBADF, 'cannot verify qualifier FD release')
                    else:
                        raise RuntimeError('qualifier FD remains open')
                result['detached'] = True
            except BaseException as error:
                result['cleanup_error'] = str(error)
        signal.alarm(0)
    result['ok'] = 'error' not in result and 'cleanup_error' not in result and result['detached']
    print(json.dumps(result, sort_keys=True), flush=True)
    return 0 if result['ok'] else 1


if __name__ == '__main__':
    raise SystemExit(main())
