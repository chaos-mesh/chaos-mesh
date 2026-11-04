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

import argparse
import os
import pathlib
import subprocess

import common


def main():
    cmd_parser = argparse.ArgumentParser(
        description='Helper script to build Chaos Mesh image.')
    cmd_parser.add_argument(
        'name',
        metavar="NAME",
        help="the name of image")
    cmd_parser.add_argument(
        'path',
        metavar="PATH",
        help="the path of the Dockerfile build directory")

    args = cmd_parser.parse_args()
    name = args.name
    image_full_name = common.get_image_full_name(name)

    if common.get_image_env(name, "BUILD") == "1":
        target_platform = common.get_target_platform()
        cmd = ["docker", "buildx", "build", "--load"]

        if os.getenv("DOCKER_CACHE") == "1":
            cache_dir = os.path.join(
                os.getenv("DOCKER_CACHE_DIR", f"{os.getcwd()}/.cache/"),
                f"image-{name}",
            )
            pathlib.Path(cache_dir).mkdir(parents=True, exist_ok=True)
            cmd += ["--cache-to", f"type=local,dest={cache_dir}"]
            if os.getenv("DISABLE_CACHE_FROM") != "1":
                cmd += ["--cache-from", f"type=local,src={cache_dir}"]

        cmd += ["--platform", f"linux/{target_platform}"]

        for env_key in common.EXPORT_ENV_VARIABLES:
            value = os.getenv(env_key)
            if value is not None:
                cmd += ["--build-arg", f"{env_key}={value}"]

        cmd += ["--build-arg", f"TARGET_PLATFORM={target_platform}"]
        cmd += ["-t", image_full_name, args.path]
    else:
        cmd = ["docker", "pull", image_full_name]

    print(" ".join(cmd))
    subprocess.run(cmd, check=True)


if __name__ == '__main__':
    main()
