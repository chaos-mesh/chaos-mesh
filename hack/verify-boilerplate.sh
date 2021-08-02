#!/usr/bin/env bash

# Copyright 2020 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(unset CDPATH && cd $(dirname "${BASH_SOURCE[0]}")/.. && pwd)
cd $ROOT

boiler="${ROOT}/hack/boilerplate/boilerplate.py"

# ignored files is a list of files we should ignore, e.g. k8s script.
# one file per line
ignored_files='./hack/cherry_pick_pull.sh
hack/generate-internal-groups.sh'

files=($(find . -type f -not \( \
        -path './hack/boilerplate/*' \
        -o -path './.git/*' \
        -o -path './.*/*' \
        -o -path './vendor/*' \
        -o -path './ui/*' \
        -o -path '*/Makefile' \
        -o -path '*/Dockerfile' \
        -o -path './images/*' \
        -o -path './pkg/uiserver/embedded_assets_handler.go' \
        -o -path '*/pb/*' \
        -o -path '*/*.deepcopy.go' \
    \) | grep -v -F "$ignored_files"
))

files_need_boilerplate=()
while IFS=$'\n' read -r line; do
  files_need_boilerplate+=( "$line" )
done < <("${boiler}" "${files[@]}")

# Run boilerplate check
if [[ ${#files_need_boilerplate[@]} -gt 0 ]]; then
    for file in "${files_need_boilerplate[@]}"; do
        echo "Boilerplate header is wrong for: ${file}" >&2
    done
    exit 1
fi
