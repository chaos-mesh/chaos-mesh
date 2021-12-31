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

set -euo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PROJECT_DIR="$(dirname "$DIR")"

# See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

cd "$PROJECT_DIR"

export GOBIN=$PROJECT_DIR/bin
export PATH=$GOBIN:$PATH

echo "+ Preflight check"
if [ ! -d "ui/app/build" ]; then
  echo "  - Error: UI assets must be built first"
  exit 1
fi

echo "+ Embed UI assets"

# OSX related
DS_Store=ui/app/build/.DS_Store
[ -f $DS_Store ] && rm $DS_Store

go run tools/assets_generate/main.go ui_server

HANDLER_PATH=pkg/dashboard/uiserver/embedded_assets_handler.go
mv assets_vfsdata.go $HANDLER_PATH
echo "  - Assets handler written to $HANDLER_PATH"
