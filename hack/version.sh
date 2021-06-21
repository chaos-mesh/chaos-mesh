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

set -euo pipefail

function chaos_mesh::version::get_version_vars() {
  if [[ -n ${GIT_COMMIT-} ]] || GIT_COMMIT=$(git rev-parse "HEAD^{commit}" 2>/dev/null); then
    # Use git describe to find the version based on tags.
    if [[ -n ${GIT_VERSION-} ]] || GIT_VERSION=$(git describe --tags --abbrev=14 "${GIT_COMMIT}^{commit}" 2>/dev/null); then
      # if current commit is not on a certain tag
      if ! git describe --tags --exact-match >/dev/null 2>&1 ; then
        # GIT_VERSION=gitBranch-gitCommitHash
        IFS='-' read -ra GIT_ARRAY <<< "$GIT_VERSION"
        GIT_VERSION=$(git rev-parse --abbrev-ref HEAD)-${GIT_ARRAY[${#GIT_ARRAY[@]}-1]}
      fi
    fi
  fi
}

function chaos_mesh::version::ldflag() {
  local key=${1}
  local val=${2}

  echo "-X 'github.com/chaos-mesh/chaos-mesh/pkg/version.${key}=${val}'"
}

# Prints the value that needs to be passed to the -ldflags parameter of go build
function chaos_mesh::version::ldflags() {
  chaos_mesh::version::get_version_vars

  local buildDate=
  [[ -z ${SOURCE_DATE_EPOCH-} ]] || buildDate="--date=@${SOURCE_DATE_EPOCH}"
  local -a ldflags=($(chaos_mesh::version::ldflag "buildDate" "$(date ${buildDate} -u +'%Y-%m-%dT%H:%M:%SZ')"))
  if [[ -n ${GIT_COMMIT-} ]]; then
    ldflags+=($(chaos_mesh::version::ldflag "gitCommit" "${GIT_COMMIT}"))
  fi

  if [[ -n ${GIT_VERSION-} ]]; then
    ldflags+=($(chaos_mesh::version::ldflag "gitVersion" "${GIT_VERSION}"))
  fi

  # The -ldflags parameter takes a single string, so join the output.
  echo "${ldflags[*]-}"
}

# output -ldflags parameters
chaos_mesh::version::ldflags
