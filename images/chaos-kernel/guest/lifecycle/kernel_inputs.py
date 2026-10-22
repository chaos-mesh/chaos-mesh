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

import argparse
import hashlib
import json
from pathlib import Path
import subprocess


def kernel_identity(bpfki):
    script = Path(__file__).resolve().parent
    lock = json.loads((script / 'inputs.lock.json').read_text())
    toolchain = '\n'.join(subprocess.check_output([tool, '--version'], timeout=60).decode().strip()
                          for tool in ['gcc', 'ld'])
    inputs = {'builder_version': 1, 'linux_sha256': lock['assets']['linux']['sha256'],
              'base_fragment_sha256': hashlib.sha256((bpfki / 'tests/kernel/vm/kernel.config').read_bytes()).hexdigest(),
              'cm_fragment_sha256': hashlib.sha256((script / 'kernel.config').read_bytes()).hexdigest(),
              'toolchain': toolchain}
    fingerprint = hashlib.sha256(json.dumps(inputs, sort_keys=True).encode()).hexdigest()
    return {'inputs': inputs, 'fingerprint': fingerprint}


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('bpfki', type=Path)
    parser.add_argument('--fingerprint', action='store_true')
    args = parser.parse_args()
    identity = kernel_identity(args.bpfki)
    print(identity['fingerprint'] if args.fingerprint else json.dumps(identity, sort_keys=True))
