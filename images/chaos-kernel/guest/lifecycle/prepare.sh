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
if [[ $# != 4 || ${GITHUB_ACTIONS:-} != true || ${RUNNER_ENVIRONMENT:-} != github-hosted ||
      ${RUNNER_OS:-} != Linux || ${RUNNER_ARCH:-} != X64 ||
      ${RUNNER_TEMP:-} != /home/runner/work/_temp || $EUID == 0 ||
      $(uname -s) != Linux || $(uname -m) != x86_64 ]]; then
  echo 'Usage on a disposable GitHub-hosted amd64 runner: prepare.sh BPFKI_CHECKOUT IMAGE_DIR RUNTIME_EVIDENCE_DIR OUTPUT_DIR' >&2
  exit 2
fi
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
python3 - "$script_dir" "$@" <<'PYTHON'
import base64
import gzip
import hashlib
import io
import json
import os
from pathlib import Path, PurePosixPath
import re
import shutil
import stat
import subprocess
import sys
import tarfile
import tempfile

script, bpfki, producer, evidence, output = map(lambda p: Path(p).resolve(), sys.argv[1:])
repo = script.parents[3]
GiB = 1024 ** 3
MiB = 1024 ** 2
release = '6.8.0-cm-kernel-ci'
need_output = Path(os.environ['RUNNER_TEMP']).resolve()
if not output.is_relative_to(need_output) or any(output.is_relative_to(path) or path.is_relative_to(output) for path in [producer, evidence, bpfki, repo]):
    raise RuntimeError('output must be a separate owned runner temporary directory')
output.mkdir(parents=True, exist_ok=True)
if any((output / name).exists() for name in ['bundle.tar.gz', 'bundle.sha256', 'review']):
    raise RuntimeError('preparation output already contains a result')
review = output / 'review'
review.mkdir()
work = Path(tempfile.mkdtemp(prefix='cm-lifecycle-', dir=os.environ['RUNNER_TEMP']))
stage = work / 'bundle'
stage.mkdir()
payload = stage / 'payload'
payload.mkdir()
context = work / 'context'
context.mkdir()
containers = []
owned_images = []


def need(ok, message):
    if not ok:
        raise RuntimeError(message)


def run(args, *, cwd=None, timeout=3600, data=None, stdout=None):
    print('+ ' + ' '.join(map(str, args)), flush=True)
    return subprocess.run(list(map(str, args)), cwd=cwd, input=data, stdout=stdout,
                          check=True, timeout=timeout)


def read(args, *, cwd=None, timeout=60):
    return subprocess.check_output(list(map(str, args)), cwd=cwd, timeout=timeout).decode().strip()


def sha(path):
    h = hashlib.sha256()
    with Path(path).open('rb') as stream:
        for block in iter(lambda: stream.read(MiB), b''):
            h.update(block)
    return h.hexdigest()


def write(path, value):
    Path(path).write_text(json.dumps(value, sort_keys=True, indent=2) + '\n')


def load(path):
    return json.loads(Path(path).read_text())


def inventory(root):
    records = {}
    for path in sorted(root.rglob('*')):
        name = path.relative_to(root).as_posix()
        mode = path.lstat().st_mode
        item = {'mode': stat.S_IMODE(mode)}
        if stat.S_ISLNK(mode):
            target = os.readlink(path)
            need(not target.startswith('/'), 'absolute share link: ' + name)
            need(path.resolve().is_relative_to(root.resolve()), 'escaping share link: ' + name)
            need(path.exists(), 'dangling share link: ' + name)
            item.update(kind='symlink', target=target)
        elif stat.S_ISDIR(mode):
            item['kind'] = 'directory'
        else:
            need(stat.S_ISREG(mode), 'special share entry: ' + name)
            item.update(kind='file', sha256=sha(path), size=path.stat().st_size)
        records[name] = item
    return records


def download(url, target, checksum=None):
    if target.exists():
        need(checksum is not None and sha(target) == checksum, 'cached input checksum mismatch: ' + str(target))
        return
    target.parent.mkdir(parents=True, exist_ok=True)
    part = target.with_name(target.name + '.part')
    run(['curl', '--fail', '--location', '--proto', '=https', '--proto-redir', '=https',
         '--retry', '3', '--max-time', '600', url, '-o', part], timeout=650)
    if checksum:
        need(sha(part) == checksum, 'download checksum mismatch: ' + url)
    part.replace(target)


def extract_file(archive, member, destination):
    with tarfile.open(archive) as source:
        candidates = [item for item in source.getmembers() if item.name.removeprefix('./') == member.removeprefix('./')]
        need(len(candidates) == 1, 'missing or ambiguous tool archive member')
        info = candidates[0]
        need(info.isfile(), 'nonregular tool archive member')
        destination.parent.mkdir(parents=True, exist_ok=True)
        with source.extractfile(info) as src, destination.open('wb') as dst:
            shutil.copyfileobj(src, dst)
    destination.chmod(0o755)


def layer_record(source, expected_diff):
    compressed = hashlib.sha256()
    raw = tempfile.TemporaryFile()
    for block in iter(lambda: source.read(MiB), b''):
        compressed.update(block)
        raw.write(block)
    size = raw.tell()
    raw.seek(0)
    magic = raw.read(2)
    raw.seek(0)
    unpacked = gzip.GzipFile(fileobj=raw) if magic == b'\x1f\x8b' else raw
    content = tempfile.TemporaryFile()
    diff = hashlib.sha256()
    for block in iter(lambda: unpacked.read(MiB), b''):
        diff.update(block)
        content.write(block)
    need('sha256:' + diff.hexdigest() == expected_diff, 'image diffID mismatch')
    content.seek(0)
    footprint = 0
    with tarfile.open(fileobj=content, mode='r:') as layer:
        for entry in layer:
            need(not PurePosixPath(entry.name).is_absolute() and '..' not in PurePosixPath(entry.name).parts,
                 'unsafe layer path')
            footprint += 4096 + ((entry.size + 4095) // 4096) * 4096
    raw.close()
    content.close()
    return {'digest': 'sha256:' + compressed.hexdigest(), 'size': size, 'diff_id': expected_diff,
            'unpacked_bytes': footprint,
            'media_type': 'application/vnd.oci.image.layer.v1.tar' + ('+gzip' if magic == b'\x1f\x8b' else '')}


def docker_archive(role, producer_tag, canonical, archive):
    with tarfile.open(archive) as saved:
        manifest_bytes = saved.extractfile('manifest.json').read()
        manifests = json.loads(manifest_bytes)
        need(len(manifests) == 1, 'expected one Docker image')
        manifest = manifests[0]
        tags = manifest['RepoTags']
        need(len(tags) == 1, 'unexpected saved image tags')
        saved_tag = tags[0]
        normalized = saved_tag if '/' in saved_tag else 'docker.io/library/' + saved_tag
        need(normalized == canonical, 'saved image does not match canonical tag')
        if role == 'kernel':
            need(saved_tag == producer_tag == 'chaos-kernel:ci', 'qualified raw producer tag changed')
        producer_tag = saved_tag
        config_bytes = saved.extractfile(manifest['Config']).read()
        config = json.loads(config_bytes)
        need(config['architecture'] == 'amd64' and config['os'] == 'linux', 'wrong image platform')
        diffs = config['rootfs']['diff_ids']
        need(len(diffs) == len(manifest['Layers']), 'layer count differs')
        layers = [layer_record(saved.extractfile(name), diff) for name, diff in zip(manifest['Layers'], diffs)]
    return {'role': role, 'producer_tag': producer_tag, 'tag': canonical,
            'config_digest': 'sha256:' + hashlib.sha256(config_bytes).hexdigest(), 'diff_ids': diffs,
            'layers': layers, 'architecture': 'amd64', 'source_sha': source_sha,
            'archive': archive.relative_to(payload).as_posix(), 'archive_sha256': sha(archive),
            'archive_manifest_sha256': hashlib.sha256(manifest_bytes).hexdigest()}


def save_image(role, tag):
    archive = payload / 'images' / (role + '.tar.gz')
    with archive.open('wb') as dst:
        docker = subprocess.Popen(['docker', 'image', 'save', tag], stdout=subprocess.PIPE)
        try:
            with gzip.GzipFile(fileobj=dst, mode='wb', compresslevel=1, mtime=0) as compressed:
                shutil.copyfileobj(docker.stdout, compressed)
            need(docker.wait(timeout=300) == 0, 'docker save failed')
        finally:
            if docker.poll() is None:
                docker.kill()
                docker.wait()
    return docker_archive(role, tag, tag, archive)


try:
    lock = load(script / 'inputs.lock.json')
    source_sha = read(['git', 'rev-parse', 'HEAD'], cwd=repo)
    need(source_sha == os.environ['GITHUB_SHA'], 'checkout differs from workflow merge SHA')
    run(['git', 'diff', '--exit-code', 'HEAD'], cwd=repo)
    need(read(['git', 'rev-parse', 'HEAD'], cwd=bpfki) == lock['bpfki_commit'], 'wrong bpfki checkout')
    run(['git', 'diff', '--exit-code', 'HEAD'], cwd=bpfki)
    provenance = load(os.environ['CM_LIFECYCLE_PROVENANCE'])
    required = ['base_repository', 'workflow_path', 'workflow_id', 'run_id', 'run_attempt', 'job_id',
                'event', 'pr_number', 'head_repository', 'head_sha', 'base_sha', 'github_sha',
                'checkout_sha', 'producer_artifact_id', 'runtime_evidence_artifact_id']
    need(all(provenance.get(key) for key in required), 'incomplete workflow provenance')
    need(provenance['base_repository'] == os.environ['GITHUB_REPOSITORY'] == 'chaos-mesh/chaos-mesh', 'wrong base repository')
    need(provenance['workflow_path'] == '.github/workflows/kernel_image.yml', 'wrong workflow')
    need(provenance['event'] == os.environ['GITHUB_EVENT_NAME'] == 'pull_request', 'wrong preparation event')
    event = load(os.environ['GITHUB_EVENT_PATH'])
    pull_request = event['pull_request']
    need(str(provenance['pr_number']) == str(event['number']) == str(pull_request['number']), 'wrong pull request')
    need(provenance['head_repository'] == pull_request['head']['repo']['full_name'] and
         provenance['head_sha'] == pull_request['head']['sha'] and
         provenance['base_sha'] == pull_request['base']['sha'], 'pull request source provenance differs')
    need(str(provenance['run_id']) == os.environ['GITHUB_RUN_ID'] and
         str(provenance['run_attempt']) == os.environ['GITHUB_RUN_ATTEMPT'], 'different preparation run')
    need(provenance['checkout_sha'] == provenance['github_sha'] == source_sha, 'wrong frozen checkout')
    write(review / 'provenance.json', provenance)
    runtime_dirs = list(evidence.rglob('image-provenance.json'))
    runtime_dirs = [p.parent for p in runtime_dirs if (p.parent / 'image-result.json').exists()]
    need(len(runtime_dirs) == 1, 'ambiguous image runtime evidence')
    runtime = runtime_dirs[0]
    gate = payload / 'gate'
    gate.mkdir()
    shutil.copytree(runtime, gate / 'runtime')
    shutil.copytree(producer, gate / 'producer', ignore=shutil.ignore_patterns('image.tar.gz'))
    preserved_result = load(gate / 'runtime/image-result.json')
    run(['python3', script.parent / 'manifest.py', 'result', gate / 'runtime'])
    need(load(gate / 'runtime/image-result.json') == preserved_result, 'raw runtime result differs')
    image_record = json.loads(read(['python3', script.parent / 'manifest.py', 'image',
                                   producer / 'image.tar.gz', producer / 'image-inspect.json']))
    need(image_record == load(producer / 'image.json') == load(runtime / 'image.json'), 'producer identity differs')
    image_provenance = load(runtime / 'image-provenance.json')
    need(image_provenance['image'] == image_record, 'runtime qualified another producer')
    need(image_provenance['runtime'] == load(runtime / 'runtime.json'), 'runtime closure differs')
    need(preserved_result['image_id'] == image_record['image_id'] and preserved_result['passed'] is True,
         'image gate did not pass')
    need(image_provenance['runtime']['build']['bcc_tag'] == 'v0.37.0' and
         image_provenance['runtime']['build']['bpfki_commit'] == lock['bpfki_commit'], 'wrong qualified source')
    environments = list((runtime / 'guest').glob('bpfki-kernel-*/environment.json'))
    need(len(environments) == 1, 'ambiguous guest environment')
    environment = load(environments[0])
    for key in ['GITHUB_RUN_ID', 'GITHUB_RUN_ATTEMPT', 'GITHUB_SHA', 'GITHUB_REPOSITORY']:
        need(environment['runner'][key] == os.environ[key], 'qualification belongs to a different run: ' + key)
    need(image_provenance['source_files'] == load(runtime / 'source-files.json'), 'source inventory differs')
    for name, record in image_provenance['source_files'].items():
        if 'sha256' in record:
            blob = subprocess.check_output(['git', '-C', str(bpfki), 'show', lock['bpfki_commit'] + ':' + name.lstrip('/')])
            need(hashlib.sha256(blob).hexdigest() == record['sha256'], 'qualified source changed')
    serial = (runtime / 'serial.log').read_text().replace('\r', '')
    begin, end = '\nBPFKI_EVIDENCE_BEGIN\n', 'BPFKI_EVIDENCE_END\n'
    need(serial.count(begin) == serial.count(end) == 1, 'incomplete qualification serial evidence')
    encoded = serial.split(begin)[1].split(end)[0]
    raw = base64.b64decode(''.join(encoded.splitlines()), validate=True)
    need('reboot: Power down' in serial.split(end)[1], 'qualification did not power down')
    seen = set()
    with tarfile.open(fileobj=io.BytesIO(raw), mode='r:gz') as archive:
        for member in archive:
            name = PurePosixPath(member.name)
            need(not name.is_absolute() and '..' not in name.parts, 'unsafe qualification evidence')
            if member.isdir():
                continue
            name = name.as_posix()
            need(member.isfile() and name not in seen, 'unexpected qualification entry')
            seen.add(name)
            need(archive.extractfile(member).read() == (runtime / 'guest' / name).read_bytes(), 'serial evidence differs')
    need(seen == {path.relative_to(runtime / 'guest').as_posix() for path in (runtime / 'guest').rglob('*') if path.is_file()},
         'raw qualification file set differs from serial archive')
    need(load(runtime / 'guest/vm-result.json')['exit_code'] == 0 and
         (runtime / 'qemu-exit-code.txt').read_text().strip() == '0', 'qualification exit mismatch')
    provenance['producer'] = image_record
    provenance['runtime_result_sha256'] = sha(runtime / 'image-result.json')
    provenance['runtime_provenance_sha256'] = sha(runtime / 'image-provenance.json')
    provenance['raw_evidence_inventory'] = inventory(gate)
    write(review / 'provenance.json', provenance)

    assets = {}
    cache = Path.home() / '.cache/chaos-mesh/lifecycle/assets'
    for name, asset in lock['assets'].items():
        target = (Path.home() / '.cache/bpfki/kernel/linux-6.8.tar.xz' if name == 'linux'
                  else cache / asset['sha256'] / asset['url'].rsplit('/', 1)[1])
        download(asset['url'], target, asset['sha256'])
        assets[name] = target
    write(review / 'inputs.lock.json', lock)
    tools = context / 'tools'
    (tools / 'usr/local/bin').mkdir(parents=True)
    for name in ['kubeadm', 'kubelet', 'kubectl']:
        shutil.copyfile(assets[name], tools / 'usr/local/bin' / name)
        (tools / 'usr/local/bin' / name).chmod(0o755)
    (tools / 'usr/local/sbin').mkdir(parents=True)
    shutil.copyfile(assets['runc'], tools / 'usr/local/sbin/runc')
    (tools / 'usr/local/sbin/runc').chmod(0o755)
    for name in ['containerd', 'containerd-shim-runc-v2', 'ctr']:
        extract_file(assets['containerd'], 'bin/' + name, tools / 'usr/local/bin' / name)
    extract_file(assets['helm'], 'linux-amd64/helm', tools / 'usr/local/bin/helm')
    extract_file(assets['crictl'], 'crictl', tools / 'usr/local/bin/crictl')
    for name in ['bridge', 'host-local', 'loopback']:
        extract_file(assets['cni'], './' + name, tools / 'opt/cni/bin' / name)
    expected_images = sorted(image['tag'] for image in lock['images'])
    actual_images = sorted(read([tools / 'usr/local/bin/kubeadm', 'config', 'images', 'list',
                                 '--kubernetes-version', 'v1.34.5']).splitlines())
    need(actual_images == expected_images, 'kubeadm image list differs from lock')
    write(review / 'kubeadm-images.json', actual_images)

    kernel_identity = json.loads(read(['python3', script / 'kernel_inputs.py', bpfki]))
    kernel_inputs = kernel_identity['inputs']
    fingerprint = kernel_identity['fingerprint']
    kernel = Path.home() / '.cache/bpfki/kernel/cm-linux-6.8' / fingerprint
    kernel.parent.mkdir(parents=True, exist_ok=True)
    with (kernel.parent / (fingerprint + '.lock')).open('w') as mutex:
        import fcntl
        fcntl.flock(mutex, fcntl.LOCK_EX)
        if not (kernel / '.cm-complete').exists():
            kernel.mkdir(exist_ok=True)
            run(['tar', '-xJf', assets['linux'], '--strip-components=1', '-C', kernel])
            os.environ.update(KBUILD_BUILD_TIMESTAMP='2024-03-10 00:00:00 UTC', KBUILD_BUILD_USER='chaos-mesh', KBUILD_BUILD_HOST='github-ci')
            run(['make', '-C', kernel, 'x86_64_defconfig'])
            run([kernel / 'scripts/kconfig/merge_config.sh', '-m', '.config',
                 bpfki / 'tests/kernel/vm/kernel.config', script / 'kernel.config'], cwd=kernel)
            run(['make', '-C', kernel, 'olddefconfig'])
            requested = {}
            for fragment in [bpfki / 'tests/kernel/vm/kernel.config', script / 'kernel.config']:
                for line in fragment.read_text().splitlines():
                    if line.startswith('CONFIG_'):
                        requested[line.split('=', 1)[0]] = line
            actual = set((kernel / '.config').read_text().splitlines())
            for line in requested.values():
                need(line in actual, 'kernel option did not survive merge: ' + line)
            run(['make', '-C', kernel, '-j2', 'bzImage'], timeout=2400)
            need(read(['make', '-s', '-C', kernel, 'kernelrelease']) == release, 'wrong kernel release')
            write(kernel / '.cm-complete', {'config_sha256': sha(kernel / '.config'),
                  'bzimage_sha256': sha(kernel / 'arch/x86/boot/bzImage'), 'fingerprint': fingerprint})
        completed = load(kernel / '.cm-complete')
        need(completed['config_sha256'] == sha(kernel / '.config') and
             completed['bzimage_sha256'] == sha(kernel / 'arch/x86/boot/bzImage'), 'cached kernel changed')
        need('"' + release + '"' in (kernel / 'include/generated/utsrelease.h').read_text(), 'generated headers mismatch')
        shutil.copyfile(kernel / 'arch/x86/boot/bzImage', stage / 'bzImage')
        headers = stage / 'headers'
        headers.mkdir()
        for name in ['include', 'arch/x86/include']:
            shutil.copytree(kernel / name, headers / name, symlinks=True)
        shutil.copyfile(kernel / '.config', headers / '.config')
        shutil.copyfile(kernel / 'Makefile', headers / 'Makefile')
        header_inventory = inventory(headers)
        headers_hash = hashlib.sha256(json.dumps(header_inventory, sort_keys=True, separators=(',', ':')).encode()).hexdigest()
    shutil.copyfile(kernel / '.config', review / 'kernel.config')
    write(review / 'kernel.json', dict(completed, release=release, inputs=kernel_inputs, headers_sha256=headers_hash))
    write(review / 'headers.json', header_inventory)

    shutil.copyfile(assets['base'], context / 'base.tar.gz')
    shutil.copytree(script, context / 'lifecycle', ignore=shutil.ignore_patterns('__pycache__'))
    shutil.copyfile(bpfki / 'src/filter_layout.h', context / 'filter_layout.h')
    shutil.copyfile(script / 'inputs.lock.json', context / 'inputs.lock.json')
    (context / 'root-overlay').mkdir()
    filter_hash = sha(context / 'filter_layout.h')
    package_seal = r'''
import hashlib
import json
from pathlib import Path
import re
import shutil
import subprocess

root = Path('/var/lib/cm-lifecycle-seal')
root.mkdir(exist_ok=True, parents=True)
lock = json.loads(Path('/opt/inputs.lock.json').read_text())

def require(ok, message):
    if not ok:
        raise RuntimeError(message)

def command(args):
    return subprocess.check_output(args, timeout=120)

def digest(data):
    return hashlib.sha256(data).hexdigest()

def records(data):
    for paragraph in data.decode().split('\n\n'):
        value = {}
        for line in paragraph.splitlines():
            if ': ' in line and not line.startswith(' '):
                key, text = line.split(': ', 1)
                value[key] = text
        if value:
            yield value

keyring = Path('/usr/share/keyrings/ubuntu-archive-keyring.gpg')
shutil.copyfile(keyring, root / 'ubuntu-archive-keyring.gpg')
trust = []
release_hashes = {}
for path in sorted(Path('/var/lib/apt/lists').glob('*InRelease')):
    checked = subprocess.run(['gpgv', '--status-fd', '1', '--keyring', str(keyring), str(path)],
                             capture_output=True, timeout=30, check=True)
    signatures = [line.split()[2:] for line in checked.stdout.decode().splitlines() if line.startswith('[GNUPG:] VALIDSIG ')]
    require(signatures and all(parts[-1] in lock['apt']['trusted_primary_fingerprints'] for parts in signatures), 'untrusted apt signer')
    text = path.read_text()
    hashes = {}
    section = text.split('\nSHA256:\n', 1)[1].split('\n\n', 1)[0]
    for line in section.splitlines():
        if not line.startswith(' '):
            break
        checksum, size, name = line.split()
        hashes[name] = {'sha256': checksum, 'size': int(size)}
    release_hashes[path.name] = hashes
    shutil.copyfile(path, root / path.name)
    trust.append({'file': path.name, 'sha256': digest(path.read_bytes()), 'signatures': signatures,
                  'verification': checked.stdout.decode(), 'diagnostics': checked.stderr.decode()})
require(len(trust) == 3, 'expected noble, noble-updates and noble-security signed releases')
indexes = []
packages = {}
for target in records(command(['apt-get', 'indextargets'])):
    if target.get('Identifier') != 'Packages':
        continue
    require(target.get('Trusted') == 'yes' and target.get('Architecture') == 'amd64', 'untrusted package index')
    file = Path(target['Filename'])
    data = command(['/usr/lib/apt/apt-helper', 'cat-file', str(file)])
    checksum = digest(data)
    matches = [name for name, hashes in release_hashes.items()
               if hashes.get(target['MetaKey']) == {'sha256': checksum, 'size': len(data)}]
    require(len(matches) == 1, 'package index not bound to one signed release')
    destination = root / file.name
    shutil.copyfile(file, destination)
    item = {'uri': target['URI'], 'meta_key': target['MetaKey'], 'filename': file.name,
            'sha256': digest(file.read_bytes()), 'uncompressed_sha256': checksum,
            'signed_release': matches[0], 'origin': target['Origin'], 'suite': target['Suite']}
    indexes.append(item)
    for pkg in records(data):
        if 'SHA256' not in pkg:
            continue
        identity = (pkg['Package'], pkg['Version'], pkg['Architecture'])
        packages.setdefault(identity, []).append({'sha256': pkg['SHA256'], 'size': int(pkg['Size']),
            'filename': pkg['Filename'], 'index': item['filename'], 'origin': target['Origin'],
            'suite': target['Suite'], 'index_uri': target['URI']})
require(indexes, 'no signed package indexes')
format_string = '${Package}\t${Version}\t${Architecture}\t${db:Status-Status}\n'
installed = command(['dpkg-query', '-W', '-f=' + format_string])
(root / 'installed-manifest.tsv').write_bytes(installed)
base = {tuple(line.split('\t')[:3]) for line in (root / 'base-manifest.tsv').read_text().splitlines()
        if line.endswith('\tinstalled')}
closure = []
debs = {}
for file in Path('/var/cache/apt/archives').glob('*.deb'):
    meta = command(['dpkg-deb', '-f', str(file), 'Package', 'Version', 'Architecture']).decode()
    values = dict(line.split(': ', 1) for line in meta.splitlines())
    debs[(values['Package'], values['Version'], values['Architecture'])] = file
for line in installed.decode().splitlines():
    package, version, arch, status = line.split('\t')
    if status != 'installed':
        continue
    require(not re.match(r'^(linux-(image|modules|headers)|kmod|containerd|runc|kubelet|kubeadm|openssh)', package),
            'unexpected installed package: ' + package)
    identity = (package, version, arch)
    if identity in base:
        closure.append({'package': package, 'version': version, 'architecture': arch, 'source': 'verified-base'})
        continue
    require(identity in debs and identity in packages, 'missing downloaded package or index metadata: ' + package)
    file = debs[identity]
    checksum = digest(file.read_bytes())
    candidates = [p for p in packages[identity] if p['sha256'] == checksum and p['size'] == file.stat().st_size]
    require(candidates, 'cached/downloaded deb mismatch: ' + str(file))
    closure.append({'package': package, 'version': version, 'architecture': arch,
                    'source': 'signed-apt-index', 'deb_sha256': checksum, 'deb_size': file.stat().st_size,
                    'matching_origins': candidates})
(root / 'package-closure.json').write_text(json.dumps({'schema_version': 1, 'base': lock['assets']['base'],
    'repositories': lock['apt']['repositories'], 'keyring_sha256': digest(keyring.read_bytes()),
    'signed_releases': trust, 'indexes': indexes, 'packages': closure}, sort_keys=True, indent=2) + '\n')
'''
    (context / 'package-seal.py').write_text(package_seal)
    suffix = source_sha[:16] + '-' + os.environ['GITHUB_RUN_ID'] + '-' + os.environ['GITHUB_RUN_ATTEMPT']
    fixture_tag = 'cm-lifecycle-fixture-build:' + suffix
    root_tag = 'cm-lifecycle-rootfs:' + suffix
    owned_images.extend([fixture_tag, root_tag])
    run(['docker', 'build', '--platform', 'linux/amd64', '--target', 'fixture', '--progress=plain',
         '--build-arg', 'FILTER_HEADER_SHA256=' + filter_hash, '-f', script / 'Dockerfile', '-t', fixture_tag, context])
    container = read(['docker', 'create', fixture_tag, '/bin/true'])
    containers.append(container)
    worker_context = work / 'worker'
    worker_context.mkdir()
    run(['docker', 'cp', container + ':/out/cm-lifecycle-fixture', worker_context / 'cm-lifecycle-fixture'])
    run(['docker', 'cp', container + ':/out/fixture-toolchain.txt', review / 'fixture-toolchain.txt'])
    run(['docker', 'rm', container])
    containers.remove(container)
    worker_hash = sha(worker_context / 'cm-lifecycle-fixture')
    shutil.copyfile(worker_context / 'cm-lifecycle-fixture', tools / 'usr/local/bin/cm-lifecycle-fixture')
    (tools / 'usr/local/bin/cm-lifecycle-fixture').chmod(0o755)
    worker_tag = 'docker.io/library/cm-lifecycle-worker:' + suffix
    owned_images.append(worker_tag)
    run(['docker', 'build', '--platform', 'linux/amd64', '--network=none', '--progress=plain',
         '-f', script / 'Dockerfile.worker', '-t', worker_tag, worker_context])
    run(['make', '-j2', 'image-chaos-mesh', 'image-chaos-daemon', 'IMAGE_TAG=' + suffix,
         'TARGET_PLATFORM=amd64'], cwd=repo, timeout=2400)
    controller_tag = 'ghcr.io/chaos-mesh/chaos-mesh:' + suffix
    daemon_tag = 'ghcr.io/chaos-mesh/chaos-daemon:' + suffix
    owned_images.extend([controller_tag, daemon_tag])
    (payload / 'images').mkdir()
    images = [save_image(role, tag) for role, tag in [('controller', controller_tag), ('daemon', daemon_tag), ('worker', worker_tag)]]
    kernel_archive = payload / 'images/kernel.tar.gz'
    shutil.copyfile(producer / 'image.tar.gz', kernel_archive)
    kernel_image = docker_archive('kernel', 'chaos-kernel:ci', 'docker.io/library/chaos-kernel:ci', kernel_archive)
    need(kernel_image['config_digest'] == image_record['image_id'] and
         kernel_image['diff_ids'] == image_record['layers'] and
         kernel_image['archive_sha256'] == image_record['archive_sha256'], 'qualified image archive changed')
    images.append(kernel_image)
    registry_cache = Path.home() / '.cache/chaos-mesh/lifecycle/registry'
    for locked in lock['images']:
        name = locked['tag'].removeprefix('registry.k8s.io/').rsplit(':', 1)[0]
        digest = locked['manifest_digest']
        manifest_path = registry_cache / digest.replace(':', '-')
        download('https://registry.k8s.io/v2/' + name + '/manifests/' + digest, manifest_path, digest.removeprefix('sha256:'))
        manifest = load(manifest_path)
        need(manifest.get('schemaVersion') == 2 and 'layers' in manifest, 'expected child image manifest')
        descriptors = [manifest['config']] + manifest['layers']
        blobs = {}
        for descriptor in descriptors:
            blob_digest = descriptor['digest']
            need(re.fullmatch(r'sha256:[0-9a-f]{64}', blob_digest) is not None, 'invalid registry digest')
            target = registry_cache / blob_digest.replace(':', '-')
            download('https://registry.k8s.io/v2/' + name + '/blobs/' + blob_digest, target, blob_digest[7:])
            need(target.stat().st_size == descriptor['size'], 'registry blob size mismatch')
            blobs[blob_digest] = target
        config = load(blobs[manifest['config']['digest']])
        need(config['architecture'] == 'amd64' and config['os'] == 'linux', 'locked child is not linux/amd64')
        diffs = config['rootfs']['diff_ids']
        need(len(diffs) == len(manifest['layers']), 'registry diffID count mismatch')
        layers = []
        for descriptor, diff in zip(manifest['layers'], diffs):
            with blobs[descriptor['digest']].open('rb') as stream:
                item = layer_record(stream, diff)
            need(item['digest'] == descriptor['digest'] and item['size'] == descriptor['size'], 'layer descriptor mismatch')
            layers.append(item)
        archive = payload / 'images' / (locked['role'] + '.tar.gz')
        with tarfile.open(archive, 'w:gz', compresslevel=1) as tar:
            def add_bytes(name, content):
                info = tarfile.TarInfo(name)
                info.size = len(content)
                info.mode = 0o644
                tar.addfile(info, io.BytesIO(content))
            add_bytes('oci-layout', b'{"imageLayoutVersion":"1.0.0"}')
            index = {'schemaVersion': 2, 'manifests': [{'mediaType': manifest['mediaType'], 'digest': digest,
                'size': manifest_path.stat().st_size, 'platform': {'os': 'linux', 'architecture': 'amd64'},
                'annotations': {'io.containerd.image.name': locked['tag'], 'org.opencontainers.image.ref.name': locked['tag']}}]}
            add_bytes('index.json', json.dumps(index).encode())
            for blob_digest, path in dict(blobs, **{digest: manifest_path}).items():
                tar.add(path, arcname='blobs/sha256/' + blob_digest[7:], recursive=False)
        images.append({'role': locked['role'], 'tag': locked['tag'], 'producer_tag': locked['tag'],
            'registry_manifest_digest': digest, 'config_digest': manifest['config']['digest'], 'diff_ids': diffs,
            'layers': layers, 'architecture': 'amd64', 'archive': archive.relative_to(payload).as_posix(),
            'archive_sha256': sha(archive), 'archive_manifest_sha256': sha(manifest_path)})
    write(review / 'images.json', images)
    shutil.copytree(repo / 'helm/chaos-mesh', payload / 'chart', ignore=shutil.ignore_patterns('.DS_Store'))
    shutil.copytree(script, payload / 'lifecycle', ignore=shutil.ignore_patterns('__pycache__'))
    node = {'schema_version': 1, 'source_sha': source_sha, 'bpfki_commit': lock['bpfki_commit'],
            'kernel': {'release': release, 'config_sha256': sha(kernel / '.config'), 'headers_sha256': headers_hash,
                       'bzimage_sha256': sha(stage / 'bzImage')},
            'images': images, 'worker': {'binary_sha256': worker_hash, 'filter_header_sha256': filter_hash}}
    write(payload / 'node.json', node)
    shutil.copyfile(payload / 'node.json', context / 'node.json')
    run(['python3', script / 'node.py', 'configure', context / 'root-overlay', context / 'node.json'])
    boot = context / 'root-overlay/boot'
    boot.mkdir(exist_ok=True, parents=True)
    shutil.copyfile(kernel / '.config', boot / ('config-' + release))
    modules = context / 'root-overlay/lib/modules' / release
    modules.mkdir(parents=True, exist_ok=True)
    for name in ['build', 'source']:
        target = modules / name
        if not target.exists() and not target.is_symlink():
            target.symlink_to('/usr/src/linux-' + release)
    with tarfile.open(assets['base']) as base_archive:
        lib_entries = [entry for entry in base_archive if entry.name.removeprefix('./').rstrip('/') == 'lib']
        need(len(lib_entries) == 1 and lib_entries[0].issym() and lib_entries[0].linkname == 'usr/lib',
             'Ubuntu base does not have the expected merged /usr layout')
    shutil.copytree(context / 'root-overlay/lib', context / 'root-overlay/usr/lib', symlinks=True, dirs_exist_ok=True)
    shutil.rmtree(context / 'root-overlay/lib')
    run(['docker', 'build', '--platform', 'linux/amd64', '--target', 'node', '--progress=plain',
         '--build-arg', 'FILTER_HEADER_SHA256=' + filter_hash, '-f', script / 'Dockerfile', '-t', root_tag, context])
    run(['git', 'diff', '--exit-code', 'HEAD'], cwd=repo)
    run(['git', 'diff', '--exit-code', 'HEAD'], cwd=bpfki)
    container = read(['docker', 'create', root_tag, '/bin/true'])
    containers.append(container)
    rootfs = work / 'rootfs'
    rootfs.mkdir()
    exported = work / 'rootfs.tar'
    with exported.open('wb') as dst:
        run(['docker', 'export', container], stdout=dst)
    exported_sha = sha(exported)
    run(['sudo', 'tar', '--numeric-owner', '-xpf', exported, '-C', rootfs])
    run(['docker', 'rm', container])
    containers.remove(container)
    exported.unlink()
    run(['sudo', 'cp', '-a', rootfs / 'var/lib/cm-lifecycle-seal', review / 'rootfs-seal'])
    run(['sudo', 'cp', '-a', rootfs / 'etc/cm-lifecycle', review / 'configuration'])
    run(['sudo', 'chown', '-R', str(os.getuid()) + ':' + str(os.getgid()), review])
    package_closure = load(review / 'rootfs-seal/package-closure.json')
    runtime_audit = load(review / 'rootfs-seal/runtime-audit.json')
    need(package_closure['packages'] and runtime_audit, 'missing package or runtime closure')
    configured = {}
    for path in sorted((context / 'root-overlay').rglob('*')):
        name = path.relative_to(context / 'root-overlay').as_posix()
        if path.is_symlink():
            configured[name] = {'symlink': os.readlink(path)}
        elif path.is_file():
            destination = review / 'generated' / name
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(path, destination)
            configured[name] = {'sha256': sha(path), 'mode': stat.S_IMODE(path.stat().st_mode)}
    write(review / 'configured-files.json', configured)
    render_code = """
import json, subprocess, yaml
raw = subprocess.check_output(['helm','template','chaos-mesh','/review-chart','--namespace','chaos-mesh',
                               '--kube-version','v1.34.5','-f','/review-values/values.json'], timeout=60, text=True)
documents = [item for item in yaml.safe_load_all(raw) if item and item.get('kind') != 'Secret']
print(json.dumps({'certificate_note':'Secrets omitted. Guest Helm generates fresh private certificate material.',
                  'documents':documents}, sort_keys=True, indent=2))
"""
    with (review / 'rendered-chart.json').open('wb') as destination:
        run(['docker', 'run', '--rm', '--network=none', '--cap-drop=ALL', '--security-opt=no-new-privileges',
             '--read-only', '--pids-limit=64', '--memory=1g', '--cpus=2',
             '--mount', 'type=bind,src=' + str(payload / 'chart') + ',dst=/review-chart,readonly',
             '--mount', 'type=bind,src=' + str(context / 'root-overlay/etc/cm-lifecycle') + ',dst=/review-values,readonly',
             '--entrypoint', 'python3', root_tag, '-c', render_code], stdout=destination, timeout=90)
    rootfs_size = int(read(['sudo', 'du', '-sx', '-B1', rootfs]).split()[0])
    apparent = int(read(['sudo', 'du', '-sx', '--apparent-size', '-B1', rootfs]).split()[0])
    unique_content, unique_snapshots = {}, {}
    for image in images:
        for layer in image['layers']:
            unique_content[layer['digest']] = layer['size']
        chain = ''
        for layer in image['layers']:
            chain = layer['diff_id'] if not chain else 'sha256:' + hashlib.sha256((chain + ' ' + layer['diff_id']).encode()).hexdigest()
            unique_snapshots[chain] = layer['unpacked_bytes']
    capacity = {'disk_bytes': 16 * GiB, 'rootfs_bytes': max(rootfs_size, apparent),
        'compressed_content_bytes': sum(unique_content.values()), 'unpacked_snapshot_bytes': sum(unique_snapshots.values()),
        'image_metadata_bytes': 128 * MiB, 'writable_snapshots_bytes': GiB, 'etcd_bytes': 512 * MiB,
        'logs_and_evidence_bytes': 256 * MiB, 'filesystem_overhead_bytes': GiB + 128 * MiB,
        'required_headroom_bytes': GiB}
    capacity['projected_guest_bytes'] = sum(value for key, value in capacity.items() if key != 'disk_bytes')
    write(review / 'capacity.json', capacity)
    need(capacity['projected_guest_bytes'] <= 16 * GiB, 'guest does not fit fixed 16 GiB disk')
    archive_growth = sum(path.stat().st_size for path in stage.rglob('*') if path.is_file()) + max(rootfs_size, apparent) + 256 * MiB
    need(shutil.disk_usage(output).free >= 20 * GiB + archive_growth,
         'host lacks full disk allocation, bundle archive growth and 4 GiB reserve')
    capacity['preparation_archive_growth_allowance_bytes'] = archive_growth
    disk = stage / 'rootfs.ext4'
    with disk.open('wb') as target:
        target.truncate(16 * GiB)
    run(['sudo', 'mkfs.ext4', '-q', '-F', '-d', rootfs, disk])
    superblock = read(['sudo', 'dumpe2fs', '-h', disk])
    (review / 'rootfs-superblock.txt').write_text(superblock + '\n')
    free = int(re.search(r'^Free blocks:\s+(\d+)', superblock, re.M)[1])
    reserved = int(re.search(r'^Reserved block count:\s+(\d+)', superblock, re.M)[1])
    block_size = int(re.search(r'^Block size:\s+(\d+)', superblock, re.M)[1])
    future_usage = capacity['projected_guest_bytes'] - capacity['rootfs_bytes'] - capacity['filesystem_overhead_bytes']
    need((free - reserved) * block_size >= future_usage, 'formatted disk lacks projected runtime space')
    capacity.update(pristine_disk_sha256=sha(disk), apparent_disk_bytes=disk.stat().st_size,
                    allocated_disk_bytes=disk.stat().st_blocks * 512, exported_filesystem_sha256=exported_sha,
                    usable_free_bytes=(free - reserved) * block_size)
    write(review / 'capacity.json', capacity)
    shutil.copytree(review, stage / 'review')
    envelope = {'cpus': 2, 'memory_mib': 6144, 'disk_bytes': 16 * GiB, 'guest_deadline_seconds': 840,
                'outer_deadline_seconds': 900, 'termination_grace_seconds': 20, 'nic': 'none',
                'serial_and_evidence_limit_bytes': 8 * MiB, 'read_only_shares': ['headers', 'payload']}
    run(['git', 'diff', '--exit-code', 'HEAD'], cwd=repo)
    run(['git', 'diff', '--exit-code', 'HEAD'], cwd=bpfki)
    need(read(['git', 'rev-parse', 'HEAD'], cwd=repo) == source_sha and
         read(['git', 'rev-parse', 'HEAD'], cwd=bpfki) == lock['bpfki_commit'], 'source revision changed during preparation')
    inventory(stage / 'headers')
    inventory(stage / 'payload')
    manifest = {'schema_version': 1, 'source_sha': source_sha, 'node_sha256': sha(payload / 'node.json'),
                'provenance': provenance, 'kernel': node['kernel'], 'capacity': capacity, 'envelope': envelope,
                'files': inventory(stage)}
    write(stage / 'bundle.json', manifest)
    archive = output / 'bundle.tar.gz'
    with archive.open('wb') as dst:
        packing = subprocess.Popen(['tar', '--sparse', '--format=pax', '--numeric-owner', '-C', str(stage), '-cf', '-', '.'], stdout=subprocess.PIPE)
        try:
            with gzip.GzipFile(fileobj=dst, mode='wb', compresslevel=1, mtime=0) as compressed:
                shutil.copyfileobj(packing.stdout, compressed)
            need(packing.wait(timeout=300) == 0, 'bundle packing failed')
        finally:
            if packing.poll() is None:
                packing.kill()
                packing.wait()
    remaining_allocation = max(0, 16 * GiB - disk.stat().st_blocks * 512)
    need(shutil.disk_usage(output).free >= remaining_allocation + 4 * GiB + 9 * MiB,
         'prepared files leave insufficient host space for full guest allocation')
    bundle_sha = sha(archive)
    (output / 'bundle.sha256').write_text(bundle_sha + '  bundle.tar.gz\n')
    write(output / 'preparation.json', {'passed': True, 'bundle_sha256': bundle_sha,
         'bundle_bytes': archive.stat().st_size, 'provenance': provenance, 'capacity': capacity, 'envelope': envelope})
    print('Prepared sealed lifecycle bundle. No guest was launched.', flush=True)
except BaseException as error:
    write(output / 'preparation.json', {'passed': False, 'error': str(error)})
    for name in ['bundle.tar.gz', 'bundle.sha256']:
        (output / name).unlink(missing_ok=True)
    raise
finally:
    for container in containers:
        subprocess.run(['docker', 'rm', '-f', container], timeout=30, check=False)
    for tag in owned_images:
        subprocess.run(['docker', 'image', 'rm', tag], timeout=30, check=False)
    subprocess.run(['sudo', 'rm', '-rf', '--', str(work)], timeout=120, check=False)
PYTHON
