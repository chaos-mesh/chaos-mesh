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

# This script would report the tag of build-env and dev-env to use based on configuartion file env-images.yaml.
#
# Usage:
# On master branch:
# ./hack/env-image-tag.sh build-env, output: latest
# ./hack/env-image-tag.sh dev-env, output: latest
# On release-2.1 branch:
# ./hack/env-image-tag.sh build-env, output: release-2.1
# ./hack/env-image-tag.sh dev-env, output: release-2.1

set -euo pipefail

DIR="$( cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd )"
PROJECT_DIR="$(dirname "$DIR")"

if [ "$#" -eq  "0" ]; then
  echo "Usage: $0 <env-image-name>"
  exit 1
fi

if [[ "$1" == "dev-env" || "$1" == "build-env"  ]]; then
  docker run -i --rm mikefarah/yq:4.24.5 ".$1.tag" < "$PROJECT_DIR"/env-images.yaml
  exit 0
fi

echo "Error: $1 is not a valid env-image name, available: dev-env, build-env"
exit 1
