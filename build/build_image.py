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
build an image in the given path with the given name and environment
configuration
"""

import os
import argparse
import subprocess
import pathlib

import utils
import common


def get_image_env(name, env):
    """
    get environment variable related with an image according to the priority:
    1. IMAGE_<name>_<env>
    2. IMAGE_<env>

    the image name will be automatically converted to the unserscore_uppsercase
    format, for example, "chaos-mesh" will be converted to "CHAOS_MESH"
    """
    default_env = os.getenv("IMAGE_" + env)
    env_mid_name = utils.underscore_uppercase(name)

    env = os.getenv(f"IMAGE_{env_mid_name}_{env}", default_env)
    if env == "":
        env = default_env
    return env


def get_image_tag(name):
    """
    get the tag of the image
    """
    return get_image_env(name, "TAG")


def get_image_build(name):
    """
    get whether this image should be built
    """
    return get_image_env(name, "BUILD")


def get_image_full_name(name):
    """
    get the full tag of an image
    """
    tag = get_image_tag(name)
    return f"ghcr.io/chaos-mesh/{name}:{tag}"


def pass_env_to_build_arg(cmd, arg_name):
    """
    pass the environment variable to the build arguments
    """
    if os.getenv(arg_name) is not None:
        cmd += ["--build-arg", f"{arg_name}={os.getenv(arg_name)}"]


def main():
    """
    entrypoint of this script
    """
    cmd_parser = argparse.ArgumentParser(
        description='Helper script to build Chaos Mesh image.')
    cmd_parser.add_argument(
        'name',
        metavar="NAME",
        type=str,
        nargs=1,
        help="the name of image")
    cmd_parser.add_argument(
        'path',
        metavar="PATH",
        type=str,
        nargs=1,
        help="the path of the Dockerfile build directory")

    args = cmd_parser.parse_args()
    name = args.name[0]
    image_full_name = get_image_full_name(name)

    env = os.environ.copy()
    cmd = []
    if get_image_build(name) == "1":
        if os.getenv("DOCKER_CACHE") == "1":
            env.update({"DOCKER_BUILDKIT": "1",
                   "DOCKER_CLI_EXPERIMENTAL": "enabled"})
            cache_dir = os.path.join(
                os.getenv("DOCKER_CACHE_DIR", f"{os.getcwd()}/.cache/"),
                f"image-{name}"
                )
            pathlib.Path(cache_dir).mkdir(parents=True, exist_ok=True)
            cmd = [
                "docker",
                "buildx",
                "build",
                "--load",
                "--cache-to",
                f"type=local,dest={cache_dir}"]
            if os.getenv("DISABLE_CACHE_FROM") != "1":
                cmd += ["--cache-from", f"type=local,src={cache_dir}"]
        else:
            if os.getenv("TARGET_PLATFORM") is not None:
                env.update({"DOCKER_BUILDKIT": "1"})
                cmd = [
                    "docker",
                    "buildx",
                    "build",
                    "--load",
                    "--platform",
                    f"linux/{os.getenv('TARGET_PLATFORM')}"]
            else:
                # This branch is split to avoid to use `buildx`, as `buildx` is
                # not supported on some CI environment
                env.update({"DOCKER_BUILDKIT": "1"})
                cmd = ["docker", "build"]

        for env_key in common.export_env_variables:
            pass_env_to_build_arg(cmd, env_key)

        target_platform = utils.get_target_platform()
        cmd += ["--build-arg", f"TARGET_PLATFORM={target_platform.platform}"]
        cmd += ["-t", image_full_name, args.path[0]]
    else:
        cmd = ["docker", "pull", image_full_name]

    print(" ".join(cmd))
    subprocess.run(cmd, env=env, check=True)


if __name__ == '__main__':
    main()
