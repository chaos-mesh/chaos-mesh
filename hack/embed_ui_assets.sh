#!/usr/bin/env bash

# This script embeds UI assets into Golang source file. UI assets must be already built
# before calling this script.
#
# Available flags:
# NO_ASSET_BUILD_TAG=1
#   No build tags will be included in the generated source code.
# ASSET_BUILD_TAG=X
#   Customize the build tag of the generated source code. If unspecified, build tag will be "ui_server".

set -euo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PROJECT_DIR="$(dirname "$DIR")"

# See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

cd "$PROJECT_DIR"

export GOBIN=$PROJECT_DIR/bin
export PATH=$GOBIN:$PATH

if [ "${NO_ASSET_BUILD_TAG:-}" = "1" ]; then
  BUILD_TAG_PARAMETER=""
else
  BUILD_TAG_PARAMETER="-tags ${ASSET_BUILD_TAG:-ui_server}"
fi

echo "+ Preflight check"
if [ ! -d "ui/build" ]; then
  echo "  - Error: UI assets must be built first"
  exit 1
fi

echo "+ Install bindata tools"
go install github.com/elazarl/go-bindata-assetfs/go-bindata-assetfs
go install github.com/go-bindata/go-bindata/v3/go-bindata

echo "+ Clean up go mod"
go mod tidy

echo "+ Embed UI assets"

go-bindata-assetfs -pkg uiserver -prefix ui $BUILD_TAG_PARAMETER ui/build/...
HANDLER_PATH=pkg/uiserver/embedded_assets_handler.go
mv bindata_assetfs.go $HANDLER_PATH
echo "  - Assets handler written to $HANDLER_PATH"
