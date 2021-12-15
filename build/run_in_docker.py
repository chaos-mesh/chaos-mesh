#!/usr/bin/env python
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

import argparse
import os
import subprocess
import sys
import pathlib

import build_image
import common
import utils

def pass_env_to_docker_arg(cmd, arg_name):
    if os.getenv(arg_name) != None:
        cmd += ["--env", "%s=%s" % (arg_name, os.getenv(arg_name))]

if __name__ == '__main__':
    cmd = argparse.ArgumentParser(description='Helper script to run make in docker env.')
    cmd.add_argument('--interactive', action='store_true', dest='interactive', help='Run in interactive mode')
    cmd.set_defaults(interactive=False)

    cmd.add_argument('--no-check', action='store_false', dest='check', help='Check the return value and exit')
    cmd.set_defaults(check=True)

    cmd.add_argument('env_name', metavar="ENV_NAME", type=str, nargs=1, help="the name of environment image")
    cmd.add_argument('commands', metavar="COMMANDS", type=str, nargs='+', help="the commands to run in docker")

    args = cmd.parse_args()

    if os.getenv("IN_DOCKER") == "1":
        sys.exit("Already in docker, exiting")

    env_image_full_name = build_image.get_image_full_name(args.env_name[0])

    cmd = ["docker", "run", "--rm", "--privileged"]
    if args.interactive:
        cmd += ["-it"]

    for env_key in common.export_env_variables:
        pass_env_to_docker_arg(cmd, env_key)
    
    cwd = os.getcwd()
    cmd += ["--env", "IN_DOCKER=1"]
    cmd += ["--volume", "%s:%s" % (cwd, cwd)]
    cmd += ["--user", "%s:%s" % (os.getuid(), os.getgid())]
    
    target_platform = utils.get_target_platform()
    if os.getenv("TARGET_PLATFORM") != None:
        cmd += ["--platform", "linux/%s" % os.getenv("TARGET_PLATFORM")]
    if target_platform == "arm64":
        cmd += ["--env", "ETCD_UNSUPPORTED_ARCH=arm64"]
    
    if os.getenv("GO_BUILD_CACHE") != None:
        tmp_go_dir = "%s/chaos-mesh-gopath" % os.getenv("GO_BUILD_CACHE")
        tmp_go_build_dir = "%s/chaos-mesh-gobuild" % os.getenv("GO_BUILD_CACHE")

        pathlib.Path(tmp_go_dir).mkdir(parents=True, exist_ok=True)
        pathlib.Path(tmp_go_build_dir).mkdir(parents=True, exist_ok=True)
        cmd += ["--volume", "%s:/tmp/go" % tmp_go_dir]
        cmd += ["--volume", "%s:/tmp/go-build" % tmp_go_build_dir]
    
    cmd += ["--workdir", cwd]
    cmd += [env_image_full_name]
    cmd += ["bash", "-c", " ".join(args.commands)]

    subprocess.run(cmd, check=args.check)