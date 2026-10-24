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

import base64
from dataclasses import dataclass
import hashlib
import io
import json
import os
from pathlib import Path
import re
import selectors
import socket
import ssl
import struct
import subprocess
import sys
import tarfile
import time
import urllib.request
import zlib
from workload import chart_values, worker_pods


PAYLOAD = Path('/opt/cm-payload')
CONFIG = Path('/etc/cm-lifecycle')
EVIDENCE = Path('/var/log/cm-lifecycle')
RELEASE = '6.8.0-cm-kernel-ci'
DEADLINE = 840.0
EVIDENCE_LIMIT = 6 * 1024 * 1024
KUBECTL = ['kubectl', '--kubeconfig=/etc/kubernetes/admin.conf', '--request-timeout=10s']
CTR = ['ctr', '--address', '/run/containerd/containerd.sock', '--namespace', 'k8s.io']
ROLES = {'controller', 'daemon', 'kernel', 'worker', 'kube-apiserver',
         'kube-controller-manager', 'kube-scheduler', 'kube-proxy', 'pause', 'etcd', 'coredns'}


def require(condition, message):
    if not condition:
        raise ValueError(message)


def sha256(path):
    digest = hashlib.sha256()
    with Path(path).open('rb') as source:
        for block in iter(lambda: source.read(1024 * 1024), b''):
            digest.update(block)
    return digest.hexdigest()


def save(path, value):
    Path(path).parent.mkdir(parents=True, exist_ok=True)
    Path(path).write_text(json.dumps(value, indent=2, sort_keys=True) + '\n')


def boot_time():
    return time.clock_gettime(time.CLOCK_BOOTTIME)


@dataclass(frozen=True)
class NodeBundle:
    value: dict
    images: dict

    @classmethod
    def read(cls, path):
        value = json.loads(Path(path).read_text())
        require(value['schema_version'] == 1, 'unsupported node bundle')
        require(re.fullmatch('[0-9a-f]{40}', value['source_sha']), 'invalid source identity')
        require(value['kernel']['release'] == RELEASE, 'unexpected kernel release')
        records = value['images']
        images = {record['role']: record for record in records}
        require(set(images) == ROLES and len(records) == len(ROLES), 'image set differs')
        require(len({record['tag'] for record in records}) == len(records), 'duplicate image tag')
        for record in records:
            require(re.fullmatch('sha256:[0-9a-f]{64}', record['config_digest']), 'invalid image config')
            require(record['diff_ids'] and all(re.fullmatch('sha256:[0-9a-f]{64}', item)
                                             for item in record['diff_ids']), 'invalid image layers')
            archive = Path(record['archive'])
            require(not archive.is_absolute() and '..' not in archive.parts and archive.parts[0] == 'images',
                    'archive outside finite payload')
            require(re.fullmatch('[0-9a-f]{64}', record['archive_sha256']), 'invalid image archive hash')
            require('/' in record['tag'] and ':' in record['tag'].rsplit('/', 1)[1], 'image tag is not canonical')
        require(images['kernel']['tag'] == 'docker.io/library/chaos-kernel:ci', 'qualified kernel alias differs')
        require(images['kernel']['producer_tag'] == 'chaos-kernel:ci', 'qualified producer tag differs')
        return cls(value, images)


def configure(root, manifest):
    root = Path(root)
    require(root != Path('/'), 'configure requires a staging root')
    node = NodeBundle.read(manifest)
    identity = sha256(manifest)
    files = {}
    files['etc/hostname'] = 'cm-kernel-ci\n'
    files['etc/hosts'] = '127.0.0.1 localhost\n192.0.2.1 cm-kernel-ci\n::1 localhost\n'
    files['etc/resolv.conf'] = 'nameserver 192.0.2.53\noptions attempts:1 timeout:1\n'
    files['etc/fstab'] = ('cm-payload /opt/cm-payload 9p trans=virtio,version=9p2000.L,ro,nosuid,nodev 0 0\n'
                          f'cm-headers /usr/src/linux-{RELEASE} 9p trans=virtio,version=9p2000.L,ro,nosuid,nodev 0 0\n')
    files['etc/cm-lifecycle/guest-id'] = identity + '\n'
    files['etc/cm-lifecycle/node.json'] = Path(manifest).read_text()
    files['etc/cm-lifecycle/guest.json'] = json.dumps({'node_sha256': identity, 'source_sha': node.value['source_sha'],
                                                   'kernel_release': RELEASE}, sort_keys=True) + '\n'
    files['etc/containerd/config.toml'] = '''version = 2
root = "/var/lib/containerd"
state = "/run/containerd"
[grpc]
  address = "/run/containerd/containerd.sock"
[plugins."io.containerd.grpc.v1.cri"]
  sandbox_image = "registry.k8s.io/pause:3.10.1"
  [plugins."io.containerd.grpc.v1.cri".containerd]
    snapshotter = "overlayfs"
    default_runtime_name = "runc"
    [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
      runtime_type = "io.containerd.runc.v2"
      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
        SystemdCgroup = true
  [plugins."io.containerd.grpc.v1.cri".cni]
    bin_dir = "/opt/cni/bin"
    conf_dir = "/etc/cni/net.d"
'''
    cni = {'cniVersion': '1.0.0', 'name': 'cm-kernel', 'plugins': [
        {'type': 'bridge', 'bridge': 'cni0', 'isGateway': True, 'hairpinMode': True, 'ipMasq': False,
         'ipam': {'type': 'host-local', 'subnet': '10.244.0.0/24', 'routes': [{'dst': '0.0.0.0/0'}]}}]}
    files['etc/cni/net.d/10-cm-kernel.conflist'] = json.dumps(cni, indent=2) + '\n'
    documents = [
        {'apiVersion': 'kubeadm.k8s.io/v1beta4', 'kind': 'InitConfiguration',
         'localAPIEndpoint': {'advertiseAddress': '192.0.2.1', 'bindPort': 6443},
         'nodeRegistration': {'name': 'cm-kernel-ci', 'criSocket': 'unix:///run/containerd/containerd.sock',
                              'taints': [], 'imagePullPolicy': 'Never',
                              'kubeletExtraArgs': [{'name': 'node-ip', 'value': '192.0.2.1'}]},
         'timeouts': {'controlPlaneComponentHealthCheck': '180s', 'kubeletHealthCheck': '60s',
                      'kubernetesAPICall': '30s', 'etcdAPICall': '30s'},
         'patches': {'directory': '/etc/cm-lifecycle/kubeadm-patches'}},
        {'apiVersion': 'kubeadm.k8s.io/v1beta4', 'kind': 'ClusterConfiguration',
         'kubernetesVersion': 'v1.34.5', 'imageRepository': 'registry.k8s.io',
         'networking': {'podSubnet': '10.244.0.0/24', 'serviceSubnet': '10.96.0.0/16',
                       'dnsDomain': 'cluster.local'}},
        {'apiVersion': 'kubelet.config.k8s.io/v1beta1', 'kind': 'KubeletConfiguration',
         'cgroupDriver': 'systemd', 'failSwapOn': True, 'resolvConf': '/etc/resolv.conf'},
        {'apiVersion': 'kubeproxy.config.k8s.io/v1alpha1', 'kind': 'KubeProxyConfiguration',
         'mode': 'nftables', 'bindAddress': '192.0.2.1', 'clusterCIDR': '10.244.0.0/24'},
    ]
    files['etc/cm-lifecycle/kubeadm.yaml'] = '\n---\n'.join(json.dumps(item, indent=2) for item in documents) + '\n'
    for component in ['kube-apiserver', 'kube-controller-manager', 'kube-scheduler', 'etcd']:
        patch = {'apiVersion': 'v1', 'kind': 'Pod', 'metadata': {'name': component},
                 'spec': {'containers': [{'name': component, 'imagePullPolicy': 'Never'}]}}
        files[f'etc/cm-lifecycle/kubeadm-patches/{component}+strategic.yaml'] = json.dumps(patch, indent=2) + '\n'
    values = chart_values(node.images)
    files['etc/cm-lifecycle/values.json'] = json.dumps(values, indent=2) + '\n'
    pods = worker_pods(node.images['worker']['tag'])
    files['etc/cm-lifecycle/workers.json'] = json.dumps({'apiVersion': 'v1', 'kind': 'List', 'items': pods}, indent=2) + '\n'
    files['etc/systemd/system/containerd.service'] = '''[Unit]
Description=Container runtime for the fixed lifecycle guest
After=local-fs.target
[Service]
Type=notify
ExecStart=/usr/local/bin/containerd
Delegate=yes
KillMode=process
Restart=no
LimitNOFILE=1048576
TasksMax=infinity
'''
    files['etc/systemd/system/kubelet.service'] = '''[Unit]
Description=Kubelet for the fixed lifecycle guest
After=containerd.service
Requires=containerd.service
[Service]
Environment="KUBELET_KUBECONFIG_ARGS=--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf"
Environment="KUBELET_CONFIG_ARGS=--config=/var/lib/kubelet/config.yaml"
EnvironmentFile=-/var/lib/kubelet/kubeadm-flags.env
ExecStart=/usr/local/bin/kubelet $KUBELET_KUBECONFIG_ARGS $KUBELET_CONFIG_ARGS $KUBELET_KUBEADM_ARGS
Restart=always
RestartSec=2
Delegate=yes
KillMode=process
TasksMax=infinity
'''
    files['etc/systemd/system/cm-lifecycle.service'] = '''[Unit]
Description=Fixed KernelChaos lifecycle evidence
After=local-fs.target cm-lifecycle-watchdog.service
RequiresMountsFor=/opt/cm-payload /usr/src/linux-6.8.0-cm-kernel-ci
[Service]
Type=oneshot
ExecStart=/usr/bin/python3 /opt/cm-lifecycle/node.py guest
ExecStopPost=/usr/bin/python3 /opt/cm-lifecycle/node.py finish
Restart=no
TimeoutStartSec=infinity
TimeoutStopSec=30
StandardOutput=null
StandardError=null
[Install]
WantedBy=multi-user.target
'''
    files['etc/systemd/system/cm-lifecycle-watchdog.service'] = '''[Unit]
Description=Absolute boot deadline for the owned lifecycle guest
DefaultDependencies=no
After=systemd-remount-fs.service
Before=basic.target
[Service]
Type=simple
ExecStart=/usr/bin/python3 /opt/cm-lifecycle/node.py watchdog
Restart=no
[Install]
WantedBy=sysinit.target
'''
    files['etc/systemd/journald.conf.d/cm-lifecycle.conf'] = '[Journal]\nStorage=volatile\nRuntimeMaxUse=32M\nForwardToConsole=no\n'
    for relative, content in files.items():
        path = root / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        if path.is_symlink():
            path.unlink()
        path.write_text(content)
    links = {'etc/systemd/system/systemd-modules-load.service': '/dev/null',
             'etc/systemd/system/kmod-static-nodes.service': '/dev/null',
             'etc/systemd/system/modprobe@.service': '/dev/null',
             'etc/udev/rules.d/80-drivers.rules': '/dev/null',
             'etc/udev/rules.d/80-debian-compat.rules': '/dev/null',
             'etc/systemd/system/multi-user.target.wants/cm-lifecycle.service': '../cm-lifecycle.service',
             'etc/systemd/system/sysinit.target.wants/cm-lifecycle-watchdog.service': '../cm-lifecycle-watchdog.service',
             f'lib/modules/{RELEASE}/build': f'/usr/src/linux-{RELEASE}',
             f'lib/modules/{RELEASE}/source': f'/usr/src/linux-{RELEASE}'}
    for relative, target in links.items():
        path = root / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        if path.exists() or path.is_symlink():
            path.unlink()
        path.symlink_to(target)
    for relative in ['opt/cm-payload', f'usr/src/linux-{RELEASE}', 'run/cm-lifecycle',
                     'var/log/cm-lifecycle', 'var/lib/cm-lifecycle', 'var/lib/containerd',
                     'var/lib/kubelet', 'boot', 'proc', 'sys', 'dev']:
        (root / relative).mkdir(parents=True, exist_ok=True)
    save(root / 'etc/cm-lifecycle/configuration.json',
         {'files': {name: sha256(root / name) for name in sorted(files)}, 'symlinks': links,
          'certificate_note': 'Helm generates private certificate material in the offline guest. Review renders omit Secrets.'})


def audit_rootfs():
    destination = Path('/var/lib/cm-lifecycle-seal')
    destination.mkdir(parents=True, exist_ok=True)
    require(Path('/sbin/init').resolve() == Path('/usr/lib/systemd/systemd'), 'PID 1 is not systemd')
    required = ['/usr/lib/systemd/systemd', '/usr/bin/python3', '/usr/sbin/ip', '/usr/sbin/nft',
                '/usr/sbin/iptables-nft', '/usr/sbin/ip6tables-nft', '/usr/bin/mount', '/usr/sbin/losetup',
                '/usr/bin/cp', '/usr/bin/systemctl', '/usr/bin/journalctl', '/usr/local/bin/containerd',
                '/usr/local/bin/containerd-shim-runc-v2', '/usr/local/sbin/runc', '/usr/local/bin/kubeadm',
                '/usr/local/bin/kubelet', '/usr/local/bin/kubectl', '/usr/local/bin/ctr',
                '/usr/local/bin/crictl', '/usr/local/bin/helm', '/usr/local/bin/cm-lifecycle-fixture',
                '/opt/cni/bin/bridge', '/opt/cni/bin/host-local', '/opt/cni/bin/loopback']
    executables = {}
    for name in required:
        path = Path(name)
        require(path.is_file() and os.access(path, os.X_OK), 'missing required executable ' + name)
        run = subprocess.run(['ldd', name], stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, timeout=10)
        require('not found' not in run.stdout, 'unresolved runtime library ' + name)
        static = 'not a dynamic executable' in run.stdout or 'statically linked' in run.stdout
        require(run.returncode == 0 or static, 'ldd failed ' + name)
        dependencies = sorted(set(re.findall(r'(?:=>\s+|^\s*)(/\S+)', run.stdout, re.MULTILINE)))
        require(dependencies or static, 'no resolved dependency evidence ' + name)
        executables[name] = {'sha256': sha256(path), 'resolved': str(path.resolve()), 'static': static,
                             'ldd': run.stdout, 'libraries': {item: sha256(item) for item in dependencies}}
    masks = {'systemd-modules-load.service', 'kmod-static-nodes.service', 'modprobe@.service',
             '80-drivers.rules', '80-debian-compat.rules'}
    units = {}
    dangerous = re.compile(r'\b(modprobe|insmod|rmmod)\b|kmod\s+(?:load|insert|remove)|builtin[^\n]*kmod')
    module_unit = re.compile(r'\bmodprobe@[\w.\\:-]*\.service\b')
    roots = ['/usr/lib/systemd/system', '/usr/lib/systemd/system-generators', '/usr/lib/systemd/system-environment-generators',
             '/usr/lib/udev/rules.d', '/etc/systemd/system', '/etc/udev/rules.d', '/etc/init.d']
    for directory in roots:
        if not Path(directory).exists():
            continue
        for path in sorted(Path(directory).rglob('*')):
            if path.is_symlink():
                units[str(path)] = {'symlink': os.readlink(path)}
                continue
            if not path.is_file():
                continue
            data = path.read_bytes()
            text = data.decode('utf-8', errors='replace')
            matches = [line for line in text.splitlines()
                       if not line.lstrip().startswith('#') and dangerous.search(module_unit.sub('', line))]
            units[str(path)] = {'sha256': hashlib.sha256(data).hexdigest(), 'module_commands': matches}
            if matches:
                require(path.name in masks, 'unaudited module-loading command in ' + str(path))
    for name in masks:
        location = Path('/etc/udev/rules.d') / name if name.endswith('.rules') else Path('/etc/systemd/system') / name
        require(location.is_symlink() and os.readlink(location) == '/dev/null', 'missing module-loading mask ' + name)
    require(not list(Path('/lib/modules').rglob('*.ko*')), 'rootfs contains loadable modules')
    installed = subprocess.check_output(['dpkg-query', '-W', '-f=${Package}\t${Version}\t${Architecture}\t${db:Status-Status}\n'], text=True, timeout=30)
    require(not any(re.match(r'(linux-image|linux-modules|linux-headers|kmod)\t', line) for line in installed.splitlines()),
            'forbidden kernel/module package')
    (destination / 'installed-manifest.tsv').write_text(installed)
    save(destination / 'runtime-audit.json', {'executables': executables, 'module_files': units,
                                           'masked_module_paths': sorted(masks), 'no_loadable_modules': True})


@dataclass
class Stage:
    name: str
    end: float

    @classmethod
    def start(cls, name, seconds):
        return cls(name, min(DEADLINE, boot_time() + seconds))

    def remaining(self, cap=None):
        seconds = min(self.end, DEADLINE) - boot_time()
        require(seconds > 0, self.name + ' deadline exceeded')
        return min(seconds, cap) if cap is not None else seconds

    def command(self, args, *, data=None, stdin=None, check=True):
        result = subprocess.run(args, input=data, stdin=stdin, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                                timeout=self.remaining(), text=True)
        require(len(result.stdout) + len(result.stderr) <= 1024 * 1024, 'command output exceeds evidence limit')
        if check and result.returncode:
            failure = {'stage': self.name, 'command': args, 'returncode': result.returncode,
                       'stderr': redact(result.stderr[-65536:])}
            if args[:2] != ['helm', 'template']:
                failure['stdout'] = redact(result.stdout[-65536:])
            evidence('failed-command.json', failure)
            raise ValueError(self.name + ' command failed: ' + args[0] + ' exit ' + str(result.returncode))
        return result

    def kubectl(self, *args, data=None, check=True):
        return self.command(KUBECTL + list(args), data=data, check=check)

    def pause(self):
        time.sleep(min(.5, self.remaining()))


def evidence(name, value):
    save(EVIDENCE / name, value)
    require(sum(path.stat().st_size for path in EVIDENCE.rglob('*') if path.is_file()) <= EVIDENCE_LIMIT,
            'required evidence exceeds limit')


def module_state():
    state = {'modules': Path('/proc/modules').read_text(),
             'modules_disabled': Path('/proc/sys/kernel/modules_disabled').read_text().strip(),
             'modprobe': Path('/proc/sys/kernel/modprobe').read_text().strip()}
    require(state == {'modules': '', 'modules_disabled': '1', 'modprobe': ''}, 'module invariant changed')
    require(not list(Path('/lib/modules').rglob('*.ko*')), 'loadable module payload appeared')
    return state


def guard_guest():
    require(os.geteuid() == 0 and os.uname().release == RELEASE, 'not the owned lifecycle guest')
    identity = (CONFIG / 'guest-id').read_text().strip()
    require(re.fullmatch('[0-9a-f]{64}', identity), 'invalid guest marker')
    commandline = Path('/proc/cmdline').read_text().split()
    require(commandline.count('cm_lifecycle=' + identity) == 1 and 'root=/dev/vda' in commandline,
            'guest kernel command line identity differs')
    require(sha256(CONFIG / 'node.json') == identity, 'embedded node manifest changed')
    return identity


def header_state(node, root=Path('/')):
    headers = root / ('usr/src/linux-' + RELEASE)
    config = root / ('boot/config-' + RELEASE)
    require(sha256(config) == node.value['kernel']['config_sha256'], 'boot config differs')
    require(sha256(headers / '.config') == node.value['kernel']['config_sha256'], 'mounted config differs')
    require('"' + RELEASE + '"' in (headers / 'include/generated/utsrelease.h').read_text(), 'header release differs')
    for name in ['build', 'source']:
        link = root / ('lib/modules/' + RELEASE + '/' + name)
        require(link.is_symlink() and os.readlink(link) == '/usr/src/linux-' + RELEASE, 'header link differs')
    return {'release': RELEASE, 'config_sha256': sha256(config),
            'utsrelease_sha256': sha256(headers / 'include/generated/utsrelease.h')}


def content_json(stage, digest):
    require(re.fullmatch('sha256:[0-9a-f]{64}', digest), 'invalid imported content digest')
    raw = stage.command(CTR + ['content', 'get', digest]).stdout.encode()
    require('sha256:' + hashlib.sha256(raw).hexdigest() == digest, 'imported JSON content digest differs')
    return json.loads(raw)


def layer_content(stage, descriptor, expected_diff):
    digest = descriptor['digest']
    require(re.fullmatch('sha256:[0-9a-f]{64}', digest), 'invalid layer digest')
    compressed = hashlib.sha256()
    uncompressed = hashlib.sha256()
    decompressor = None
    media = descriptor['mediaType']
    if media.endswith('+gzip') or media.endswith('.gzip'):
        decompressor = zlib.decompressobj(16 + zlib.MAX_WBITS)
    else:
        require(media.endswith('.tar'), 'unsupported layer encoding')
    process = subprocess.Popen(CTR + ['content', 'get', digest], stdout=subprocess.PIPE, stderr=subprocess.DEVNULL)
    selector = selectors.DefaultSelector()
    selector.register(process.stdout, selectors.EVENT_READ)
    size = 0
    expanded = 0
    try:
        while True:
            require(selector.select(stage.remaining()), 'content stream deadline exceeded')
            chunk = os.read(process.stdout.fileno(), 1024 * 1024)
            if not chunk:
                break
            compressed.update(chunk)
            size += len(chunk)
            plain = decompressor.decompress(chunk) if decompressor else chunk
            uncompressed.update(plain)
            expanded += len(plain)
            require(expanded <= 16 * 1024 ** 3, 'layer exceeds disk envelope')
        if decompressor:
            require(decompressor.eof and not decompressor.unused_data, 'incomplete or concatenated layer compression')
        require(process.wait(timeout=stage.remaining()) == 0, 'content reader failed')
        require(size == descriptor['size'] and 'sha256:' + compressed.hexdigest() == digest, 'layer bytes differ')
        require('sha256:' + uncompressed.hexdigest() == expected_diff, 'decompressed layer differs')
        return {'digest': digest, 'size': size, 'diff_id': expected_diff, 'uncompressed_bytes': expanded}
    finally:
        selector.close()
        process.stdout.close()
        if process.poll() is None:
            process.kill()
            process.wait(timeout=1)


def imported_image(stage, record):
    listing = stage.command(CTR + ['images', 'ls']).stdout.splitlines()
    rows = [line.split() for line in listing[1:] if line.split() and line.split()[0] == record['tag']]
    require(len(rows) == 1 and len(rows[0]) >= 5, 'missing canonical imported image ' + record['tag'])
    target = rows[0][2]
    manifest = content_json(stage, target)
    if 'manifests' in manifest:
        children = [item for item in manifest['manifests']
                    if item.get('platform', {}).get('architecture') == 'amd64'
                    and item.get('platform', {}).get('os') == 'linux']
        require(len(children) == 1, 'ambiguous imported image platform')
        manifest_digest = children[0]['digest']
        manifest = content_json(stage, manifest_digest)
    else:
        manifest_digest = target
    require(manifest['config']['digest'] == record['config_digest'], 'imported image config differs')
    config = content_json(stage, record['config_digest'])
    require(config['architecture'] == 'amd64' and config['os'] == 'linux', 'imported architecture differs')
    require(config['rootfs']['diff_ids'] == record['diff_ids'], 'imported diff IDs differ')
    require(len(manifest['layers']) == len(record['diff_ids']), 'imported layer count differs')
    layers = [layer_content(stage, descriptor, expected)
              for descriptor, expected in zip(manifest['layers'], record['diff_ids'])]
    parent = ''
    snapshots = []
    for diff_id in record['diff_ids']:
        chain = diff_id if not parent else 'sha256:' + hashlib.sha256((parent + ' ' + diff_id).encode()).hexdigest()
        snapshot = json.loads(stage.command(CTR + ['snapshots', '--snapshotter', 'overlayfs', 'info', chain]).stdout)
        require(snapshot['Name'] == chain and snapshot.get('Parent', '') == parent and snapshot['Kind'] == 'Committed',
                'unpacked snapshot chain differs')
        snapshots.append(snapshot)
        parent = chain
    return {'role': record['role'], 'producer_tag': record.get('producer_tag'), 'tag': record['tag'],
            'target_digest': target, 'manifest_digest': manifest_digest, 'config_digest': record['config_digest'],
            'layers': layers, 'snapshots': snapshots}


def import_images(node):
    stage = Stage.start('image import', 120)
    stage.command(['systemctl', 'start', 'containerd.service'])
    imported = []
    for record in node.images.values():
        archive = PAYLOAD / record['archive']
        require(sha256(archive) == record['archive_sha256'], 'image archive changed')
        stage.remaining()
        decompression = ['gzip', '-dc', '--', str(archive)]
        process = subprocess.Popen(decompression, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL)
        try:
            stage.command(CTR + ['images', 'import', '--platform', 'linux/amd64', '-'], stdin=process.stdout)
            process.stdout.close()
            returncode = process.wait(timeout=stage.remaining())
            if returncode:
                evidence('failed-command.json', {'stage': stage.name, 'command': decompression,
                                                 'returncode': returncode, 'stderr': 'gzip decompression failed'})
                raise ValueError('image archive decompression failed: exit ' + str(returncode))
        finally:
            process.stdout.close()
            if process.poll() is None:
                process.kill()
                process.wait(timeout=1)
        imported.append(imported_image(stage, record))
    evidence('imported-images.json', imported)


def addon_documents(text, name):
    import yaml
    documents = [document for document in yaml.safe_load_all(text) if document]
    require(name in {'kube-proxy', 'coredns'}, 'unexpected addon')
    expected = name
    allowed = {'ConfigMap', 'Service', 'ServiceAccount', 'ClusterRole', 'ClusterRoleBinding', 'Role', 'RoleBinding',
               'DaemonSet' if name == 'kube-proxy' else 'Deployment'}
    workloads = []
    for document in documents:
        require(document['kind'] in allowed, 'unexpected kubeadm addon object kind')
        metadata = document['metadata']
        require(metadata.get('namespace', 'kube-system') == 'kube-system', 'addon outside kube-system')
        require(metadata['name'] in {'kube-proxy', 'kubeadm:node-proxier', 'kube-dns', 'coredns', 'system:coredns'},
                'unexpected kubeadm addon object name')
        if document['kind'] in {'DaemonSet', 'Deployment'}:
            require(metadata['name'] == expected, 'unexpected addon workload')
            containers = document['spec']['template']['spec']['containers']
            require(len(containers) == 1 and containers[0]['name'] == name, 'unexpected addon containers')
            expected_image = {'kube-proxy': 'registry.k8s.io/kube-proxy:v1.34.5',
                              'coredns': 'registry.k8s.io/coredns/coredns:v1.12.1'}[name]
            require(containers[0]['image'] == expected_image, 'unexpected addon image')
            containers[0]['imagePullPolicy'] = 'Never'
            if name == 'coredns':
                require(document['spec']['replicas'] == 2, 'CoreDNS replica count differs')
            workloads.append(document)
        if document['kind'] == 'ConfigMap' and metadata['name'] == 'coredns':
            corefile = document['data']['Corefile']
            corefile, count = re.subn(r'(?m)^\s*forward \. /etc/resolv\.conf \{\n(?:[^\n]*\n)*?\s*\}\n', '', corefile)
            require(count == 1 and 'forward ' not in corefile, 'CoreDNS forwarding block differs')
            document['data']['Corefile'] = corefile
    require(len(workloads) == 1, 'missing addon workload')
    return documents


def pods_ready(stage, namespace, count):
    while True:
        pods = json.loads(stage.kubectl('get', 'pods', '-n', namespace, '-o', 'json').stdout)['items']
        require(len(pods) <= count, 'unexpected Pod in ' + namespace)
        ready = len(pods) == count
        for pod in pods:
            require(not pod['metadata'].get('deletionTimestamp'), 'Pod is being replaced')
            statuses = pod.get('status', {}).get('containerStatuses', [])
            require(all(item['restartCount'] == 0 for item in statuses), 'container restarted')
            require(all('terminated' not in item.get('state', {}) for item in statuses), 'container terminated')
            ready = ready and len(statuses) == len(pod['spec']['containers']) and all(item['ready'] for item in statuses)
            require(all(item['imagePullPolicy'] == 'Never' for item in pod['spec']['containers']), 'image pull policy differs')
        if ready:
            return pods
        stage.pause()


def pod_identity(pods):
    return {pod['metadata']['uid']: {'name': pod['metadata']['name'],
            'containers': [{key: item[key] for key in ['name', 'containerID', 'imageID', 'restartCount']}
                           for item in pod['status']['containerStatuses']]}
            for pod in pods}


def cluster_network(stage):
    context = ssl.create_default_context(cafile='/etc/kubernetes/pki/ca.crt')
    with urllib.request.urlopen('https://10.96.0.1/readyz', context=context, timeout=stage.remaining(5)) as response:
        require(response.status == 200 and response.read(1024) == b'ok', 'API ClusterIP readiness failed')
    query_id = 27017
    question = b''.join(bytes([len(label)]) + label.encode() for label in 'kubernetes.default.svc.cluster.local'.split('.'))
    query = struct.pack('!6H', query_id, 0x100, 1, 0, 0, 0) + question + b'\x00' + struct.pack('!HH', 1, 1)
    with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as dns:
        dns.settimeout(stage.remaining(5))
        dns.sendto(query, ('10.96.0.10', 53))
        answer, source = dns.recvfrom(4096)
    require(source == ('10.96.0.10', 53) and len(answer) >= 12, 'unexpected DNS response')
    header = struct.unpack('!6H', answer[:12])
    require(header[0] == query_id and header[1] & 0x8000 and header[1] & 15 == 0 and header[3] > 0,
            'cluster DNS lookup failed')
    require(b'\x00\x01\x00\x01' in answer and socket.inet_aton('10.96.0.1') in answer, 'cluster DNS address differs')
    return {'cluster_ip_ready': True, 'dns_query_hex': query.hex(), 'dns_response_hex': answer.hex(),
            'routes': stage.command(['ip', '-json', 'route', 'show']).stdout,
            'links': stage.command(['ip', '-json', 'address', 'show']).stdout,
            'nft_rules': stage.command(['nft', '-j', 'list', 'ruleset']).stdout}


def bootstrap(node):
    stage = Stage.start('node bootstrap', 240)
    require(stage.command(['stat', '-fc', '%T', '/sys/fs/cgroup']).stdout.strip() == 'cgroup2fs', 'not unified cgroup v2')
    require(not any(' - cgroup ' in line for line in Path('/proc/self/mountinfo').read_text().splitlines()), 'cgroup v1 mount')
    require({'cpu', 'cpuset', 'io', 'memory', 'pids'} <= set(Path('/sys/fs/cgroup/cgroup.controllers').read_text().split()),
            'missing cgroup controllers')
    require(len(Path('/proc/swaps').read_text().splitlines()) == 1, 'swap is active')
    for command in [
        ['ip', 'link', 'set', 'lo', 'up'], ['ip', 'link', 'add', 'dummy0', 'type', 'dummy'],
        ['ip', 'addr', 'add', '192.0.2.1/24', 'dev', 'dummy0'], ['ip', 'link', 'set', 'dummy0', 'up'],
        ['ip', 'route', 'add', '10.96.0.0/16', 'dev', 'dummy0', 'src', '192.0.2.1'],
        ['sysctl', '-w', 'net.ipv4.ip_forward=1'], ['sysctl', '-w', 'net.bridge.bridge-nf-call-iptables=1'],
        ['update-alternatives', '--set', 'iptables', '/usr/sbin/iptables-nft'],
        ['update-alternatives', '--set', 'ip6tables', '/usr/sbin/ip6tables-nft']]:
        stage.command(command)
    require('nf_tables' in stage.command(['iptables', '--version']).stdout, 'iptables backend differs')
    require('nf_tables' in stage.command(['ip6tables', '--version']).stdout, 'ip6tables backend differs')
    stage.command(['systemctl', 'start', 'kubelet.service'])
    stage.command(['kubeadm', 'init', '--config', str(CONFIG / 'kubeadm.yaml'), '--skip-phases=addon', '--skip-token-print'])
    for name in ['kube-proxy', 'coredns']:
        rendered = stage.command(['kubeadm', 'init', 'phase', 'addon', name, '--config', str(CONFIG / 'kubeadm.yaml'),
                                  '--print-manifest']).stdout
        documents = addon_documents(rendered, name)
        evidence('addon-' + name + '.json', documents)
        stage.kubectl('apply', '-f', '-', data=json.dumps({'apiVersion': 'v1', 'kind': 'List', 'items': documents}))
    stage.kubectl('wait', 'node/cm-kernel-ci', '--for=condition=Ready', '--timeout=' + str(int(stage.remaining())) + 's')
    nodes = json.loads(stage.kubectl('get', 'nodes', '-o', 'json').stdout)['items']
    require(len(nodes) == 1 and nodes[0]['metadata']['name'] == 'cm-kernel-ci', 'unexpected node set')
    require(not nodes[0]['spec'].get('taints'), 'unexpected scheduling taint')
    system_pods = pods_ready(stage, 'kube-system', 7)
    names = [pod['metadata']['labels'].get('component', pod['metadata']['labels'].get('k8s-app')) for pod in system_pods]
    require(sorted(names) == sorted(['kube-apiserver', 'kube-controller-manager', 'kube-scheduler', 'etcd', 'kube-proxy', 'kube-dns', 'kube-dns']),
            'ordinary node component set differs')
    evidence('node-ready.json', {'node': nodes[0], 'pods': system_pods, 'module_state': module_state(),
                               'network': cluster_network(stage)})
    stage.kubectl('create', 'namespace', 'cm-kernel-ci')
    return pod_identity(system_pods)


def quantity(value, cpu=False):
    value = str(value)
    if cpu:
        return float(value[:-1]) / 1000 if value.endswith('m') else float(value)
    suffixes = {'Ki': 1024, 'Mi': 1024 ** 2, 'Gi': 1024 ** 3, 'Ti': 1024 ** 4,
                'K': 1000, 'M': 1000 ** 2, 'G': 1000 ** 3}
    for suffix, multiplier in suffixes.items():
        if value.endswith(suffix):
            return float(value[:-len(suffix)]) * multiplier
    return float(value)


def chart_requests(documents, pods, node):
    resources = []
    for document in documents:
        if document and document.get('kind') in {'Deployment', 'DaemonSet'}:
            template = document['spec']['template']['spec']
            require(not template.get('initContainers'), 'unexpected chart init container')
            resources.extend(container.get('resources', {}) for container in template['containers'])
    require(len(resources) == 3, 'chart container set differs')
    resources += [container.get('resources', {}) for pod in pods for container in pod['spec']['containers']]
    resources += [{'requests': {'cpu': '10m', 'memory': '16Mi'}}] * 2
    totals = {}
    for key in ['cpu', 'memory']:
        totals[key] = sum(quantity(resource.get('requests', {}).get(key, resource.get('limits', {}).get(key, '0')), key == 'cpu')
                          for resource in resources)
        require(0 < totals[key] < quantity(node['status']['allocatable'][key], key == 'cpu'), 'requests exceed allocatable ' + key)
    return totals


def webhook_control(stage):
    document = {'apiVersion': 'chaos-mesh.org/v1alpha1', 'kind': 'KernelChaos',
                'metadata': {'name': 'admission-check', 'namespace': 'cm-kernel-ci'},
                'spec': {'mode': 'one', 'selector': {'namespaces': ['cm-kernel-ci'], 'pods': {'cm-kernel-ci': ['target']}},
                         'containerNames': ['worker'], 'duration': 'not-a-duration',
                         'failKernRequest': {'failtype': 0, 'probability': 100, 'times': 0, 'callchain': [{'predicate': 'true'}]}}}
    invalid = stage.kubectl('create', '--dry-run=server', '-f', '-', '-o', 'json', data=json.dumps(document), check=False)
    require(invalid.returncode != 0 and 'admission webhook \"vkernelchaos.kb.io\"' in invalid.stderr and 'parse duration field error' in invalid.stderr,
            'invalid duration was not rejected by the named webhook')
    absent = stage.kubectl('get', 'kernelchaos', 'admission-check', '-n', 'cm-kernel-ci', '-o', 'json', check=False)
    require(absent.returncode != 0 and 'NotFound' in absent.stderr, 'dry-run persisted a resource')
    document['spec']['duration'] = '120s'
    del document['spec']['selector']['namespaces']
    valid = json.loads(stage.kubectl('create', '--dry-run=server', '-f', '-', '-o', 'json', data=json.dumps(document)).stdout)
    require(valid['spec']['duration'] == '120s' and valid['spec']['selector']['namespaces'] == ['cm-kernel-ci'],
            'valid admission dry-run did not default the namespace')
    webhooks = json.loads(stage.kubectl('get', 'validatingwebhookconfigurations', 'chaos-mesh-validation', '-o', 'json').stdout)
    require(all(item.get('failurePolicy') == 'Fail' for item in webhooks['webhooks']), 'webhook failure policy differs')
    evidence('admission.json', {'invalid_stderr': invalid.stderr, 'valid': valid, 'configuration': webhooks})


def chart_ready(node):
    import yaml
    stage = Stage.start('chart readiness', 120)
    rendered = stage.command(['helm', 'template', 'chaos-mesh', str(PAYLOAD / 'chart'), '-n', 'chaos-mesh',
                              '--kube-version', 'v1.34.5', '-f', str(CONFIG / 'values.json')]).stdout
    documents = [item for item in yaml.safe_load_all(rendered) if item]
    pods = json.loads(stage.kubectl('get', 'pods', '-n', 'kube-system', '-o', 'json').stdout)['items']
    kube_node = json.loads(stage.kubectl('get', 'node', 'cm-kernel-ci', '-o', 'json').stdout)
    totals = chart_requests(documents, pods, kube_node)
    review = [item for item in documents if item['kind'] != 'Secret']
    evidence('chart-review.json', {'objects': review, 'requests': totals,
                                 'certificate_note': 'The install generates new normal Helm certificates inside this guest.'})
    stage.command(['helm', 'install', 'chaos-mesh', str(PAYLOAD / 'chart'), '--namespace', 'chaos-mesh',
                   '--create-namespace', '-f', str(CONFIG / 'values.json'), '--wait', '--timeout', str(int(stage.remaining())) + 's'])
    pods = pods_ready(stage, 'chaos-mesh', 2)
    require(sorted(container['name'] for pod in pods for container in pod['spec']['containers']) == ['bpfki', 'chaos-daemon', 'chaos-mesh'],
            'chart containers differ')
    for pod in pods:
        require(not pod['spec'].get('hostNetwork'), 'chart bypassed Pod networking')
        for container in pod['spec']['containers']:
            role = {'bpfki': 'kernel', 'chaos-daemon': 'daemon', 'chaos-mesh': 'controller'}[container['name']]
            require(container['image'] == node.images[role]['tag'], 'chart uses different image')
    daemon = next(pod for pod in pods if any(item['name'] == 'bpfki' for item in pod['spec']['containers']))
    require(daemon['spec']['hostPID'] is True, 'daemon lacks host PID namespace')
    endpoints = json.loads(stage.kubectl('get', 'endpointslices', '-n', 'chaos-mesh', '-l',
                                        'kubernetes.io/service-name=chaos-daemon', '-o', 'json').stdout)
    require(any(daemon['status']['podIP'] in endpoint['addresses'] and endpoint.get('conditions', {}).get('ready') is True
                for item in endpoints['items'] for endpoint in item['endpoints']), 'daemon EndpointSlice is not ready')
    webhook_control(stage)
    evidence('chart-ready.json', {'pods': pods, 'daemon_endpoints': endpoints, 'network': cluster_network(stage)})


def initial_namespaces(stage):
    initial = os.stat('/proc/1/ns/pid').st_ino
    processes = {}
    for service in ['containerd', 'kubelet']:
        pid = int(stage.command(['systemctl', 'show', '--property=MainPID', '--value', service + '.service']).stdout)
        require(pid > 1 and os.stat(f'/proc/{pid}/ns/pid').st_ino == initial, 'runtime PID namespace differs')
        processes[service] = {'pid': pid, 'pid_namespace': initial}
    return {'init_pid_namespace': initial, 'services': processes}


def redact(text):
    text = re.sub(r'-----BEGIN [^-]+-----.*?-----END [^-]+-----', '[redacted PEM]', text, flags=re.DOTALL)
    text = re.sub(r'\b[a-z0-9]{6}\.[a-z0-9]{16}\b', '[redacted bootstrap token]', text)
    text = re.sub(r'\beyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b', '[redacted token]', text)
    text = re.sub(r'(?im)((?:bearer|token|client-key-data|certificate-authority-data)\s*[:= ]+)\S+', r'\1[redacted]', text)
    return text


def failure_diagnostics(stage):
    captures = {}
    commands = {
        'journal': ['journalctl', '-b', '-u', 'containerd', '-u', 'kubelet', '-u', 'cm-lifecycle',
                    '--no-pager', '-n', '120', '-o', 'short-monotonic'],
        'pods': KUBECTL + ['get', 'pods', '--all-namespaces', '-o', 'json'],
        'events': KUBECTL + ['get', 'events', '-n', 'chaos-mesh', '-o', 'json'],
        'nft': ['nft', '-j', 'list', 'ruleset'],
    }
    for name, command in commands.items():
        try:
            result = subprocess.run(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True,
                                    timeout=stage.remaining(2))
            captures[name] = {'returncode': result.returncode, 'stdout': redact(result.stdout[-65536:]),
                              'stderr': redact(result.stderr[-8192:]), 'bounded_diagnostic': True}
        except Exception as error:
            captures[name] = {'error': type(error).__name__ + ': ' + str(error)}
    evidence('failure-diagnostics.json', captures)


def guest():
    guard_guest()
    EVIDENCE.mkdir(parents=True, exist_ok=True)
    Path('/run/cm-lifecycle').mkdir(parents=True, exist_ok=True)
    node = NodeBundle.read(CONFIG / 'node.json')
    result = {'schema_version': 1, 'ok': False, 'cases': [], 'cleanup': {'ok': False}, 'error': ''}
    observer = None
    context = {'bundle': node.value, 'deadline': DEADLINE, 'evidence_dir': EVIDENCE}
    import verify
    try:
        require(sha256(PAYLOAD / 'node.json') == (CONFIG / 'guest-id').read_text().strip(), 'payload identity differs')
        evidence('boot.json', {'boot_id': Path('/proc/sys/kernel/random/boot_id').read_text().strip(),
                               'boot_seconds': boot_time(), 'module_state': module_state(), 'headers': header_state(node),
                               'source_sha': node.value['source_sha']})
        for mount in ['/opt/cm-payload', '/usr/src/linux-' + RELEASE]:
            rows = [row for row in Path('/proc/self/mountinfo').read_text().splitlines() if row.split()[4] == mount]
            require(len(rows) == 1 and 'ro' in rows[0].split()[5].split(',') and ' - 9p ' in rows[0], 'input share is not read-only')
        import_images(node)
        original_system = bootstrap(node)
        chart_ready(node)
        fixture = Stage.start('fixture readiness', 30)
        context['fixture_deadline'] = fixture.end
        fixture.command(['/usr/local/bin/cm-lifecycle-fixture', 'prepare'])
        fixture.kubectl('apply', '-f', str(CONFIG / 'workers.json'))
        workers = pods_ready(fixture, 'cm-kernel-ci', 2)
        evidence('fixture-ready.json', {'pods': workers, 'namespaces': initial_namespaces(fixture), 'module_state': module_state()})
        result, observer = verify.run(context)
        if result['ok']:
            ending = Stage.start('evidence and cleanup', 30)
            context['cleanup_deadline'] = ending.end
            current_system = pods_ready(ending, 'kube-system', 7)
            require(pod_identity(current_system) == original_system, 'node components changed during lifecycle')
            evidence('final-module-state.json', module_state())
    except Exception as error:
        result['ok'] = False
        result['error'] = type(error).__name__ + ': ' + str(error)
    finally:
        context.setdefault('cleanup_deadline', min(DEADLINE, boot_time() + 30))
        try:
            cleanup = verify.cleanup(context, observer or context.get('_lifecycle_observer'))
            result['cleanup'] = cleanup
            result['ok'] = result['ok'] and cleanup['ok']
            evidence('cleanup-module-state.json', module_state())
        except Exception as error:
            result['ok'] = False
            result['cleanup'] = {'ok': False, 'error': str(error)}
        if not result['ok']:
            try:
                failure_diagnostics(Stage('evidence and cleanup', context['cleanup_deadline']))
            except Exception as error:
                result['diagnostics_error'] = type(error).__name__ + ': ' + str(error)
        result['finished_boot_seconds'] = boot_time()
        evidence('guest-result.json', result)
    return 0 if result['ok'] else 1


def finish():
    guard_guest()
    EVIDENCE.mkdir(parents=True, exist_ok=True)
    termination = {'exit_code': os.environ.get('EXIT_CODE', ''), 'exit_status': os.environ.get('EXIT_STATUS', ''),
                   'boot_seconds': boot_time()}
    termination['ok'] = termination['exit_code'] == 'exited' and termination['exit_status'] == '0'
    save(EVIDENCE / 'service-exit.json', termination)
    ok = False
    try:
        paths = sorted(EVIDENCE.rglob('*'))
        require(all(path.is_file() and not path.is_symlink() for path in paths), 'unexpected evidence entry')
        require(sum(path.stat().st_size for path in paths) <= EVIDENCE_LIMIT, 'required evidence exceeds limit')
        result = json.loads((EVIDENCE / 'guest-result.json').read_text())
        ok = result['ok'] is True and result['cleanup']['ok'] is True and termination['ok'] and boot_time() < DEADLINE
        output = io.BytesIO()
        with tarfile.open(fileobj=output, mode='w:gz') as archive:
            for path in paths:
                archive.add(path, arcname=str(path.relative_to(EVIDENCE)), recursive=False)
        raw = output.getvalue()
        require(len(raw) <= EVIDENCE_LIMIT, 'compressed evidence exceeds limit')
        with open('/dev/ttyS0', 'wb', buffering=0) as serial:
            serial.write(('\nCM_LIFECYCLE_EVIDENCE_BEGIN ' + hashlib.sha256(raw).hexdigest() + ' ' + str(len(raw)) + '\n').encode())
            encoded = base64.b64encode(raw)
            for offset in range(0, len(encoded), 4096):
                serial.write(encoded[offset:offset + 4096] + b'\n')
            serial.write(b'CM_LIFECYCLE_EVIDENCE_END\n')
            serial.write(('CM_LIFECYCLE_EXIT=' + str(0 if ok else 1) + '\n').encode())
    except Exception:
        with open('/dev/ttyS0', 'wb', buffering=0) as serial:
            serial.write(b'\nCM_LIFECYCLE_EVIDENCE_INVALID\nCM_LIFECYCLE_EXIT=1\n')
    finally:
        subprocess.run(['systemctl', 'poweroff', '--no-block'], timeout=3, check=False)


def watchdog():
    guard_guest()
    while True:
        remaining = DEADLINE - boot_time()
        if remaining <= 0:
            break
        time.sleep(min(1, remaining))
    with open('/dev/ttyS0', 'w') as serial:
        serial.write('\nCM_LIFECYCLE_WATCHDOG_EXPIRED\n')
    subprocess.run(['/sbin/poweroff', '--force', '--force'], timeout=3, check=False)
    return 1


def main():
    mode, *args = sys.argv[1:]
    if mode == 'configure' and len(args) == 2:
        configure(*args)
    elif mode == 'audit-rootfs' and not args:
        audit_rootfs()
    elif mode == 'guest' and not args:
        return guest()
    elif mode == 'finish' and not args:
        finish()
    elif mode == 'watchdog' and not args:
        return watchdog()
    else:
        raise ValueError('expected configure ROOT NODE_JSON, audit-rootfs, guest, finish, or watchdog')
    return 0


if __name__ == '__main__':
    sys.exit(main())
