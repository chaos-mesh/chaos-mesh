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
functions used in multiple scripts
"""

import os
import sys
from collections import namedtuple


def underscore_uppercase(name):
    """
    convert the given name to the underscore_uppercase format
    """
    return name.replace('-', '_').upper()


def get_target_platform():
    """
    get the target platform according to the `TARGET_PLATFORM` variable or the
    `uname` syscall
    """
    Platform = namedtuple('Platform', ['platform', 'from_env'])

    if os.getenv("TARGET_PLATFORM") is not None and os.getenv("TARGET_PLATFORM") != "":
        return Platform(os.getenv("TARGET_PLATFORM"), True)

    machine = os.uname().machine
    if machine == "x86_64":
        return Platform("amd64", False)

    if machine == "amd64":
        return Platform("amd64", False)

    if machine == "arm64":
        return Platform("arm64", False)

    if machine == "aarch64":
        return Platform("arm64", False)

    sys.exit("Please run this script on amd64 or arm64 machines.")
