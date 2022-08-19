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

# Don't run this script directly, use `yarn bootstrap` to exec it.

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in
    --compact)
      COMPACT=true
      shift
      ;;
  esac
done

# step1
if [[ ! -d node_modules ]]; then
  echo "No node_modules found. Install by yarn:"

  yarn
else
  echo "Already install dependencies."
fi

# step2
echo "Build packages..."

yarn workspace @ui/mui-extends build

# step3
CHAOS_DASHBOARD_BIN=../images/chaos-dashboard/bin

if [[ "$COMPACT" == true ]]; then
  echo "--compact: skip building chaos-dashboard."
elif [[ ! -f $CHAOS_DASHBOARD_BIN/chaos-dashboard ]]; then
  echo "No chaos-dashboard binary found. Install by IN_DOCKER=1 make images/chaos-dashboard/bin/chaos-dashboard:"

  cd ..
  IN_DOCKER=1 make images/chaos-dashboard/bin/chaos-dashboard
else
  echo "Already build chaos-dashboard."
fi
