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


"""Build same-checkout components and fixed fixtures using existing image targets."""
import argparse
from dataclasses import asdict
import gzip
import json
import os
from pathlib import Path
import shutil
import subprocess
import tarfile
import urllib.request

from identity import Workflow, archive_identity, need, sha

REPO = Path(__file__).resolve().parents[2]
LIFECYCLE = REPO / 'images/chaos-kernel/guest/lifecycle'


def run(args, **kwargs):
    subprocess.run([str(arg) for arg in args], check=True, timeout=2400, **kwargs)


def download_tools(output):
    lock = json.loads((LIFECYCLE / 'inputs.lock.json').read_text())
    cache = Path.home() / '.cache/chaos-mesh/lifecycle/assets'
    cache.mkdir(parents=True, exist_ok=True)
    for name, member, executable in [('kubectl', None, 'kubectl'), ('helm', 'linux-amd64/helm', 'helm'),
                                      ('crictl', 'crictl', 'crictl'), ('containerd', 'bin/ctr', 'ctr')]:
        asset = lock['assets'][name]
        path = cache / asset['sha256']
        if not path.exists():
            partial = path.with_suffix('.partial')
            try:
                with urllib.request.urlopen(asset['url'], timeout=60) as response, partial.open('wb') as stream:
                    shutil.copyfileobj(response, stream)
                need(sha(partial) == asset['sha256'], 'download checksum mismatch')
                partial.replace(path)
            finally:
                partial.unlink(missing_ok=True)
        need(sha(path) == asset['sha256'], 'cached asset checksum mismatch')
        if member:
            with tarfile.open(path) as archive, (output / executable).open('wb') as stream:
                shutil.copyfileobj(archive.extractfile(member), stream)
        else:
            shutil.copyfile(path, output / executable)
        (output / executable).chmod(0o755)


def produce(kernel, output):
    workflow = Workflow.current()
    need(subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip() == workflow.checkout,
         'producer checkout differs from workflow')
    need(json.loads((kernel / 'workflow.json').read_text()) == asdict(workflow), 'kernel producer identity differs')
    kernel_record = json.loads((kernel / 'image.json').read_text())
    kernel_image = archive_identity(kernel / 'image.tar.gz')
    need(kernel_image['config_digest'] == kernel_record['image_id'] and
         kernel_image['diff_ids'] == kernel_record['layers'] and kernel_image['archive_sha256'] == kernel_record['archive_sha256'],
         'kernel artifact differs from checked producer image')
    output.mkdir(parents=True, exist_ok=False)
    suffix = f'{workflow.checkout[:16]}-{workflow.run_id}-{workflow.attempt}'
    run(['docker', 'load', '-i', kernel / 'image.tar.gz'])
    compiler = ['docker', 'run', '--rm', '--network=none', '--user', f'{os.getuid()}:{os.getgid()}',
                '-v', str(LIFECYCLE) + ':/src:ro', '-v', str(output) + ':/out',
                '--entrypoint', '/bin/sh', kernel_image['config_digest'], '-ec']
    run(compiler + ['''header=$(sha256sum /bpfki/src/filter_layout.h | cut -d ' ' -f 1)
cp /bpfki/src/filter_layout.h /out/filter_layout.h
g++ --version > /out/fixture-toolchain.txt
for guard in SEALED MINIKUBE; do
  g++ -std=c++17 -O2 -static -Wall -Wextra -Werror -I /bpfki/src \\
    "-DCM_GUEST_${guard}=1" "-DCM_FILTER_HEADER_SHA256=\\\"${header}\\\"" \\
    /src/fixture.cc -o "/out/fixture-${guard}"
done
mv /out/fixture-MINIKUBE /out/cm-lifecycle-fixture
rm /out/fixture-SEALED
'''])
    worker = 'docker.io/library/cm-lifecycle-worker:' + suffix
    run(['docker', 'build', '--platform', 'linux/amd64', '--network=none',
         '-f', LIFECYCLE / 'Dockerfile.worker', '-t', worker, output])
    run(['make', '-j2', 'image-chaos-mesh', 'image-chaos-daemon', 'IMAGE_TAG=' + suffix,
         'TARGET_PLATFORM=amd64'], cwd=REPO)
    images = {'kernel': dict(kernel_image, tag='docker.io/library/chaos-kernel:ci', archive='kernel.tar.gz')}
    shutil.copyfile(kernel / 'image.tar.gz', output / 'kernel.tar.gz')
    for role, tag in [('worker', worker), ('controller', 'ghcr.io/chaos-mesh/chaos-mesh:' + suffix),
                      ('daemon', 'ghcr.io/chaos-mesh/chaos-daemon:' + suffix)]:
        path = output / (role + '.tar.gz')
        with path.open('wb') as raw, gzip.GzipFile(fileobj=raw, mode='wb', compresslevel=1) as compressed:
            process = subprocess.Popen(['docker', 'save', tag], stdout=subprocess.PIPE)
            try:
                shutil.copyfileobj(process.stdout, compressed)
                need(process.wait(timeout=60) == 0, 'docker save failed')
            finally:
                process.stdout.close()
                if process.poll() is None:
                    process.kill()
                    process.wait()
        images[role] = dict(archive_identity(path), tag=tag, archive=role + '.tar.gz')
    download_tools(output)
    scripts = [LIFECYCLE / (name + '.py') for name in ('checks', 'cases', 'workload')]
    scripts += [Path(__file__).parent / (name + '.py') for name in ('identity', 'api', 'guest', 'qualify_bpf')]
    for path in scripts:
        shutil.copyfile(path, output / path.name)
    files = {p.name: sha(p) for p in output.iterdir() if p.is_file() and not p.name.endswith('.tar.gz')}
    value = {'workflow': asdict(workflow), 'images': images, 'files': files,
             'worker': {'binary_sha256': sha(output / 'cm-lifecycle-fixture'),
                        'filter_header_sha256': sha(output / 'filter_layout.h')}}
    (output / 'inputs.json').write_text(json.dumps(value, indent=2, sort_keys=True) + '\n')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('kernel', type=Path)
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    produce(args.kernel, args.output)
