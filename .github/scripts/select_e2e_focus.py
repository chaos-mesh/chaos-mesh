#!/usr/bin/env python3
# Copyright 2026 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import json
import os

groups = [
    ("iochaos", "[Graceful-Shutdown] [IOChaos]"),
    ("httpchaos", "[Graceful-Shutdown] [HTTPChaos]"),
    ("podchaos", "[Basic] [PodChaos]"),
    ("timechaos", "[Basic] [TimeChaos]"),
    ("iochaos", "[Basic] [IOChaos]"),
    ("httpchaos", "[Basic] [HTTPChaos]"),
    ("sidecar", "[Basic] [Sidecar Config]"),
    ("networkchaos", "[Basic] [NetworkChaos]"),
    ("dnschaos", "[Basic] [DNSChaos]"),
    ("stresschaos", "[Basic] [StressChaos]"),
    ("jvmchaos", "[Basic] [JVMChaos]"),
]

dedicated_controllers = {
    "controllers/podhttpchaos/": "httpchaos",
    "controllers/podnetworkchaos/": "networkchaos",
}

dedicated_daemon_files = {
    "pkg/chaosdaemon/dns_server.go": "dnschaos",
    "pkg/chaosdaemon/httpchaos_server.go": "httpchaos",
    "pkg/chaosdaemon/jvm_server.go": "jvmchaos",
    "pkg/chaosdaemon/stress_server_linux.go": "stresschaos",
}

def file_scope(path):
    shared_io_manager = "controllers/chaosimpl/iochaos/podiochaosmanager/"
    if not isinstance(path, str) or path.startswith(shared_io_manager):
        return None
    for prefix, scope in dedicated_controllers.items():
        if path.startswith(prefix):
            return scope
    for scope, _ in groups:
        if scope != "sidecar" and path.startswith(f"controllers/chaosimpl/{scope}/"):
            return scope
        if scope != "jvmchaos" and path.startswith(f"e2e-test/e2e/chaos/{scope}/"):
            return scope
    for scope in ("podchaos", "networkchaos", "jvmchaos"):
        if path in (f"e2e-test/e2e-gherkin/features/{scope}.feature",
                    f"e2e-test/e2e-gherkin/steps/{scope}_steps.go"):
            return scope
    return dedicated_daemon_files.get(path)

try:
    files = json.loads(os.environ["E2E_FILES"])
except json.JSONDecodeError:
    files = []
scopes = {file_scope(path) for path in files} if isinstance(files, list) else set()
focus = [label for scope, label in groups if not scopes or None in scopes or scope in scopes]
with open(os.environ["GITHUB_OUTPUT"], "a") as output:
    print(f"focus={json.dumps(focus)}", file=output)
