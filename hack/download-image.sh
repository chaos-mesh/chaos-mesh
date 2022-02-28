#!/usr/bin/env bash

set -e

function usage() {
    cat <<'EOF'
This command downloads and imports the chaos-mesh image from the github artifacts

Usage: hack/download-image.sh

    -h show this message and exit
    -r the github repository which running the
    -i download the artifact related with this action run id

EOF
}

function download_image() {
    local github_repository=$1
    local github_run_id=$2

    mkdir -p .cache/

    local ARTIFACT_URL=$(curl \
        -H "Accept: application/vnd.github.v3+json" \
        https://api.github.com/repos/$github_repository/actions/runs/$github_run_id/artifacts 2>/dev/null |\
        jq -r ".artifacts[0].archive_download_url")
    local TOKEN=$(echo url=https://github.com/$github_repository|\
        gh auth git-credential get|\
        grep password|\
        cut -b 10-)

    curl -L \
        -H "Accept: application/vnd.github.v3+json" \
        -H "Authorization: token $TOKEN" \
        $ARTIFACT_URL > .cache/chaos-mesh-images.zip
    unzip -o -d .cache/ .cache/chaos-mesh-images.zip
    
    for IMAGE in "chaos-mesh" "chaos-daemon" "chaos-dashboard"
    do
        docker image import .cache/$IMAGE.tar ghcr.io/chaos-mesh/$IMAGE:latest &>/dev/null
        echo "Image ghcr.io/chaos-mesh/$IMAGE:latest loaded"
    done
}

function check_executable_exists() {
    while [[ $# -gt 0 ]]; do
        local executable=$1
        if ! command -v $executable >/dev/null 2>&1; then
            echo "Error: $executable is not installed"
            exit 1
        fi

        shift
    done
}

GITHUB_REPOSITORY=""
GITHUB_ACTION_ID=""

if [ $# -eq 0 ]
  then
    usage
    exit 1
fi

check_executable_exists gh jq unzip curl

while [[ $# -gt 0 ]]; do
    case $1 in
        -r)
            GITHUB_REPOSITORY=$2
            shift
            shift
            ;;
        -i)
            GITHUB_ACTION_ID=$2
            shift
            shift
            ;;
        -h)
            usage
            exit 0
            ;;
        *)
            echo "unknown flag or option $1"
            usage
            exit 1
            ;;
    esac
done

download_image $GITHUB_REPOSITORY $GITHUB_ACTION_ID