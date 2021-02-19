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

# This script embeds UI assets into Golang source file. UI assets must be already built
# before calling this script.
#
# Available flags:
#
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

echo "+ Preflight check"
if [ ! -d "ui/build" ]; then
  echo "  - Error: UI assets must be built first"
  exit 1
fi

if [ "${NO_ASSET_BUILD_TAG:-}" = "1" ]; then
  BUILD_TAG_PARAMETER=""
else
  BUILD_TAG_PARAMETER=${ASSET_BUILD_TAG:-ui_server}
fi

echo "+ Embed UI assets"

# OSX related
DS_Store=ui/build/.DS_Store
[ -f $DS_Store ] && rm $DS_Store

go run tools/assets_generate/main.go $BUILD_TAG_PARAMETER

HANDLER_PATH=pkg/uiserver/embedded_assets_handler.go
mv assets_vfsdata.go $HANDLER_PATH
echo "  - Assets handler written to $HANDLER_PATH"
