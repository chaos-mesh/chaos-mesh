# Copyright 2021 Chaos Mesh Authors.
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

import os
import sys


EXPORT_ENV_VARIABLES = (
    "HTTP_PROXY",
    "HTTPS_PROXY",
    "GOPROXY",
    "UI",
    "LDFLAGS",
    "CRATES_MIRROR",
    "GO_BUILD_CACHE",
)

HOST_PLATFORM_ALIASES = {
    "x86_64": "amd64",
    "amd64": "amd64",
    "arm64": "arm64",
    "aarch64": "arm64",
}


def get_image_env(name, env):
    """Prefer a nonempty IMAGE_<NAME>_<ENV> over IMAGE_<ENV>."""
    default_env = os.getenv("IMAGE_" + env)
    env_mid_name = name.replace('-', '_').upper()
    return os.getenv(f"IMAGE_{env_mid_name}_{env}", default_env) or default_env


def get_image_full_name(name):
    tag = get_image_env(name, "TAG")
    return f"ghcr.io/chaos-mesh/{name}:{tag}"


def get_target_platform():
    target = os.getenv("TARGET_PLATFORM")
    if target:
        return target

    platform = HOST_PLATFORM_ALIASES.get(os.uname().machine)
    if platform:
        return platform

    sys.exit("Please run this script on amd64 or arm64 machines.")
