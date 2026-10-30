#!/usr/bin/env python3
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

"""
print the `docker run` command to run `Make` commands in a container
the arguments of `docker run` is based on the environment variables
"""

import argparse
import os
import pathlib
import shlex

import common


def main():
    cmd_parser = argparse.ArgumentParser(
        description='Helper script to run make in docker env.')
    cmd_parser.add_argument(
        '--interactive',
        action='store_true',
        help='Run in interactive mode')

    cmd_parser.add_argument(
        'env_name',
        metavar="ENV_NAME",
        help="the name of environment image")

    args = cmd_parser.parse_args()

    if os.path.exists("/.dockerenv"):
        print("bash")
        return

    env_image_full_name = common.get_image_full_name(args.env_name)

    cmd = ["docker", "run", "--rm", "--privileged"]
    if args.interactive:
        cmd += ["-it"]

    cwd = os.getcwd()
    cmd += ["--volume", f"{cwd}:{cwd}"]
    # HACK: if you have permission issues when using Docker Desktop for Mac, you can try removing this line to see if it works.
    cmd += ["--user", f"{os.getuid()}:{os.getgid()}"]

    target_platform = common.get_target_platform()
    cmd += ["--platform", f"linux/{target_platform}"]
    cmd += ["--env", f"TARGET_PLATFORM={target_platform}"]

    # For testing on arm64.
    # See https://etcd.io/docs/v3.5/op-guide/supported-platform/#unsupported-platforms.
    if target_platform == "arm64":
        cmd += ["--env", "ETCD_UNSUPPORTED_ARCH=arm64"]

    go_build_cache = os.getenv("GO_BUILD_CACHE")
    if go_build_cache:
        for directory, mount in (
            ("chaos-mesh-gopath", "/tmp/go"),
            ("chaos-mesh-gobuild", "/tmp/go-build"),
        ):
            cache_dir = f"{go_build_cache}/{directory}"
            pathlib.Path(cache_dir).mkdir(parents=True, exist_ok=True)
            cmd += ["--volume", f"{cache_dir}:{mount}"]

    for env_key in common.EXPORT_ENV_VARIABLES:
        value = os.getenv(env_key)
        if value:
            cmd += ["-e", env_key]

    cmd += ["--workdir", cwd]
    cmd += [env_image_full_name]
    cmd += ["/bin/bash"]

    print(shlex.join(cmd))


if __name__ == '__main__':
    main()
