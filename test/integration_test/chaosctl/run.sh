#!/usr/bin/env bash
# Copyright 2022 Chaos Mesh Authors.
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

test_dir=test/integration_test
pwd=`pwd`

function run() {
    script=$1
    echo "Running test $script..."
    TEST_NAME="$(basename "$(dirname "$script")")" \
    PATH="$pwd/$test_dir/utils:$PATH" \
    bash +x "$script"
}

scripts=(debug.sh recover.sh)

for name in $scripts; do
    script="$test_dir/chaosctl/$name"
    echo "run $script"
    run $script
done
