#!/usr/bin/env python

# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import os
import argparse
import subprocess

import utils
import common

def get_image_env(name, env, default):
    default_env = os.getenv("IMAGE_" + env, default)
    env_mid_name = utils.underscore_uppercase(name)
    return os.getenv("IMAGE_%s_%s" % (env_mid_name, env), default_env)

def get_image_project(name):
    return get_image_env(name, "PROJECT", "pingcap")

def get_image_registry(name):
    return get_image_env(name, "REGISTRY", "localhost:5000")

def get_image_tag(name):
    return get_image_env(name, "TAG", "latest")

def get_image_build(name):
    return get_image_env(name, "BUILD", "1")

def get_image_full_name(name):
    project = get_image_project(name)
    registry = get_image_registry(name)
    tag = get_image_tag(name)

    return "%s/%s/%s:%s" % (registry, project, name, tag)

def pass_env_to_build_arg(cmd, arg_name):
    if os.getenv(arg_name) != None:
        cmd += ["--build-arg", "%s=%s" % (arg_name, os.getenv(arg_name))]

if __name__ == '__main__':
    cmd = argparse.ArgumentParser(description='Helper script to build Chaos Mesh image.')
    cmd.add_argument('name', metavar="NAME", type=str, nargs=1, help="the name of image")
    cmd.add_argument('path', metavar="PATH", type=str, nargs=1, help="the path of the Dockerfile build directory")

    args = cmd.parse_args()
    name = args.name[0]
    image_full_name = get_image_full_name(name)

    env = {}
    cmd = []
    if get_image_build(name) == "1":
        if os.getenv("DOCKER_CACHE") == "1":
            env = {"DOCKER_BUILDKIT": "1", "DOCKER_CLI_EXPERIMENTAL": "enabled"}
            cache_dir = os.getenv("DOCKER_CACHE_DIR", "%s/.cache/image-%s" % (os.getcwd(), name))
            cmd = ["docker", "buildx", "build", "--load", "--cache-to", "type=local,dest=%s" % cache_dir]
            if os.getenv("DISABLE_CACHE_FROM") != "1":
                cmd += ["--cache-from", "type=local,src=%s" % cache_dir]
        else:
            if os.getenv("TARGET_PLATFORM") != None:
                env = {"DOCKER_BUILDKIT": "1"}
                cmd = ["docker", "buildx", "build", "--load", "--platform", os.getenv("TARGET_PLATFORM")]
            else:
                # This branch is split to avoid to use `buildx`, as `buildx` is not supported on some CI environment
                env = {"DOCKER_BUILDKIT": "1"}
                cmd = ["docker", "build"]

        for env_key in common.export_env_variables:
            pass_env_to_build_arg(cmd, env_key)

        target_platform = utils.get_target_platform()
        cmd += ["--build-arg", "%s=%s" % ("TARGET_PLATFORM", target_platform)]
        if os.getenv("TARGET_PLATFORM") != None:
            cmd += ["--platform", "linux/%s" % os.getenv("TARGET_PLATFORM")]

        cmd += ["-t", image_full_name, args.path[0]]
    else:
        cmd = ["docker", "pull", image_full_name]

    print(cmd)
    # subprocess.run(cmd, env=env)