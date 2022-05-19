#!/usr/bin/env bash
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

set -eu

test_dir=test/integration_test
pwd=`pwd`

function run() {
    script=$1
    echo "Running test $script..."
    TEST_NAME="$(basename "$(dirname "$script")")" \
    PATH="$pwd/$test_dir/utils:$PATH" \
    bash +x "$script"
}

if [ "$#" -ge 1 ]; then
    test_case=$1
    if [ "$test_case" != "*" ]; then
        if [ ! -d "test/integration_test/$test_case" ]; then
            echo $test_case "not exist"
            exit 0
        fi
    fi
else
    test_case="*"
fi

if [ "$test_case" == "*" ]; then
    for script in $test_dir/$test_case/run.sh; do
        # jvmchaos and chaosd are not supported in aarch64
        # TODO: support JVMChaos / chaosd in aarch64, and remove this check
        if [[ ($script == *"jvm"* || $script == *"physical_machine"*) && "$(uname -m)" == "aarch64" ]]; then
            continue
        fi

        run $script
    done
else
    for name in $test_case; do
        script="$test_dir/$name/run.sh"
        echo "run $script"
        run $script
    done
fi