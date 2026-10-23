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
import signal
import struct
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
        config = gzip.decompress(Path('/proc/config.gz').read_bytes())
        options = dict(line.decode().split('=', 1) for line in config.splitlines() if line.startswith(b'CONFIG_') and b'=' in line)
        for option in ('BPF', 'BPF_SYSCALL', 'KPROBES', 'BPF_EVENTS', 'KPROBE_EVENTS',
                       'BPF_KPROBE_OVERRIDE', 'FUNCTION_ERROR_INJECTION', 'IKHEADERS'):
            need(options.get('CONFIG_' + option) == 'y', 'missing config ' + option)
        result['config_sha256'] = hashlib.sha256(config).hexdigest()
        result['optional_config'] = {name: options.get('CONFIG_' + name, 'absent')
                                     for name in ('FAILSLAB', 'FAULT_INJECTION')}
        allowlist = Path('/sys/kernel/debug/error_injection/list').read_text()
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
}''', debug=4)
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
