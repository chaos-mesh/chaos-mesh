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