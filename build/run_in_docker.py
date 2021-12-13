#!/usr/bin/env python

import argparse
import os
import subprocess
import sys

import build_image
import common
import utils

def pass_env_to_docker_arg(cmd, arg_name):
    cmd += ["--env", "%s=%s" % (arg_name, os.getenv(arg_name, ""))]

if __name__ == '__main__':
    cmd = argparse.ArgumentParser(description='Helper script to run make in docker env.')
    cmd.add_argument('env_name', metavar="ENV_NAME", type=str, nargs=1, help="the name of environment image")
    cmd.add_argument('commands', metavar="COMMANDS", type=str, nargs='+', help="the commands to run in docker")

    args = cmd.parse_args()

    if os.getenv("IN_DOCKER") == "1":
        sys.exit("Already in docker, exiting")

    env_image_full_name = build_image.get_image_full_name(args.env_name[0])

    cmd = ["docker", "run", "-it", "--rm"]
    for env_key in common.export_env_variables:
        pass_env_to_docker_arg(cmd, env_key)
    cmd += ["--env", "IN_DOCKER=1"]
    cmd += ["--volume", "%s:/mnt" % os.getcwd()]
    cmd += ["--user", "%s:%s" % (os.getuid(), os.getgid())]
    
    target_platform = utils.get_target_platform()
    if os.getenv("TARGET_PLATFORM") != None:
        cmd += ["--platform", "linux/%s" % os.getenv("TARGET_PLATFORM")]
    if target_platform == "arm64":
        cmd += ["--env", "ETCD_UNSUPPORTED_ARCH=arm64"]
    
    if os.getenv("GO_BUILD_CACHE") != None:
        cmd += ["--volume", "%s/chaos-mesh-gopath:/tmp/go" % os.getenv("GO_BUILD_CACHE")]
        cmd += ["--volume", "%s/chaos-mesh-gobuild:/tmp/go-build" % os.getenv("GO_BUILD_CACHE")]
    
    cmd += ["--workdir", "/mnt"]
    cmd += [env_image_full_name]
    cmd += ["bash", "-c", " ".join(args.commands)]
    subprocess.run(cmd, check=True)