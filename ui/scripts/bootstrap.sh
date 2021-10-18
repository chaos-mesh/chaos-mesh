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
#
# Don't run this script directly, use `yarn bootstrap` to exec it.

# Check node deps.
if [[ ! -d node_modules ]]; then
  echo "No node_modules found. Install by yarn:"

  yarn
else
  echo "Already install node deps."
fi

CHAOS_DASHBOARD_BIN=../images/chaos-dashboard/bin

# Check dashboard server.
if [[ ! -f $CHAOS_DASHBOARD_BIN/chaos-dashboard ]]; then
  echo "No chaos-dashboard binary found. Install by make IN_DOCKER=1 images/chaos-dashboard/bin/chaos-dashboard:"

  cd ..
  make IN_DOCKER=1 SWAGGER=1 images/chaos-dashboard/bin/chaos-dashboard && rm -f docs/docs.go && GO111MODULE=on go mod tidy
  cd -
else
  echo "Already build chaos-dashboard."
fi
