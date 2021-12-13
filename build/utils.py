import os
import sys

def underscore_uppercase(name):
    return name.replace('-', '_').upper()

def get_target_platform():
    if os.getenv("TARGET_PLATFORM") != None:
        return os.getenv("TARGET_PLATFORM")
    else:
        machine = os.uname().machine
        if machine == "x86_64":
            return "amd64"
        elif machine == "amd64":
            return "amd64"
        elif machine == "arm64":
            return "arm64"
        elif machine == "aarch64":
            return "arm64"
        else:
            sys.exit("Please run this script on amd64 or arm64 machine")