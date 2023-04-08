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
import sys
import pathlib

import build_image
import common
import utils


def pass_env_to_docker_arg(cmd, arg_name):
    """
    pass environment variable to the docker run command
    """
    if os.getenv(arg_name) is not None and os.getenv(arg_name) != "":
        cmd += ["-e", f"{arg_name}"]


def main():
    """
    entrypoint of this script
    """
    cmd_parser = argparse.ArgumentParser(
        description='Helper script to run make in docker env.')
    cmd_parser.add_argument(
        '--interactive',
        action='store_true',
        dest='interactive',
        help='Run in interactive mode')
    cmd_parser.set_defaults(interactive=False)

    cmd_parser.add_argument(
        '--no-check',
        action='store_false',
        dest='check',
        help='Check the return value and exit')
    cmd_parser.set_defaults(check=True)

    cmd_parser.add_argument(
        'env_name',
        metavar="ENV_NAME",
        type=str,
        nargs=1,
        help="the name of environment image")

    args = cmd_parser.parse_args()

    if os.path.exists("/.dockerenv"):
        print("bash")
        sys.exit(0)

    env_image_full_name = build_image.get_image_full_name(args.env_name[0])

    cmd = ["docker", "run", "--rm", "--privileged"]
    if args.interactive:
        cmd += ["-it"]

    cwd = os.getcwd()
    cmd += ["--volume", f"{cwd}:{cwd}"]
    cmd += ["--user", f"{os.getuid()}:{os.getgid()}"]

    target_platform = utils.get_target_platform()
    # if the environment variable is not set, don't pass `--platform` argument,
    # as it's not supported on some docker build environment.
    if os.getenv("TARGET_PLATFORM") is not None and os.getenv(
            "TARGET_PLATFORM") != "":
        cmd += ["--platform", f"linux/{os.getenv('TARGET_PLATFORM')}"]
    else:
        cmd += ["--env", f"TARGET_PLATFORM={target_platform}"]

    if target_platform == "arm64":
        cmd += ["--env", "ETCD_UNSUPPORTED_ARCH=arm64"]

    if os.getenv("GO_BUILD_CACHE") is not None and os.getenv(
            "GO_BUILD_CACHE") != "":
        tmp_go_dir = f"{os.getenv('GO_BUILD_CACHE')}/chaos-mesh-gopath"
        tmp_go_build_dir = f"{os.getenv('GO_BUILD_CACHE')}/chaos-mesh-gobuild"

        pathlib.Path(tmp_go_dir).mkdir(parents=True, exist_ok=True)
        pathlib.Path(tmp_go_build_dir).mkdir(parents=True, exist_ok=True)
        cmd += ["--volume", f"{tmp_go_dir}:/tmp/go"]
        cmd += ["--volume", f"{tmp_go_build_dir}:/tmp/go-build"]

    for env_key in common.export_env_variables:
        pass_env_to_docker_arg(cmd, env_key)

    cmd += ["--workdir", cwd]
    cmd += [env_image_full_name]
    cmd += ["/bin/bash"]

    print(" ".join(cmd))


if __name__ == '__main__':
    main()
