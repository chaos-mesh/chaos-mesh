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

# This script should work with the github action in `.github/workflows/upload_image_pr.yml`

set -e

function usage() {
    cat <<'EOF'
This command downloads and imports the chaos-mesh image from the github artifacts

Usage: hack/download-image.sh -r REPOSITORY -i RUN_ID

    -h show this message and exit
    -r the github repository which running the github actions
    -i download the artifact related with this action run id

EOF
}

function download_image() (
    local github_repository=$1
    local github_run_id=$2
    local download_dir
    download_dir=$(mktemp -d)
    trap 'rm -rf "$download_dir"' EXIT

    gh run download "$github_run_id" --repo "$github_repository" \
        --name chaos-mesh-images --dir "$download_dir"

    for IMAGE in "chaos-mesh" "chaos-daemon" "chaos-dashboard"
    do
        docker image load --input "$download_dir/$IMAGE.tar" &>/dev/null
        echo "Image ghcr.io/chaos-mesh/$IMAGE:latest loaded"
    done
)

function check_executable_exists() {
    while [[ $# -gt 0 ]]; do
        local executable=$1
        if ! command -v "$executable" >/dev/null 2>&1; then
            echo "Error: $executable is not installed"
            exit 1
        fi

        shift
    done
}

GITHUB_REPOSITORY=""
GITHUB_ACTION_ID=""

while getopts ':r:i:h' option; do
    case $option in
        r|i)
            if [[ -z $OPTARG || $OPTARG == -* ]]; then
                echo "Error: -$option requires a value"
                exit 1
            fi
            if [[ $option == r ]]; then
                GITHUB_REPOSITORY=$OPTARG
            else
                GITHUB_ACTION_ID=$OPTARG
            fi
            ;;
        h)
            usage
            exit 0
            ;;
        :)
            echo "Error: -$OPTARG requires a value"
            exit 1
            ;;
        *)
            echo "unknown flag or option -$OPTARG"
            usage
            exit 1
            ;;
    esac
done
shift "$((OPTIND - 1))"

if [[ -z $GITHUB_REPOSITORY || -z $GITHUB_ACTION_ID || $# -ne 0 ]]; then
    usage
    exit 1
fi

check_executable_exists gh docker

download_image "$GITHUB_REPOSITORY" "$GITHUB_ACTION_ID"
