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

# This script generates API client from the swagger annotation in the Golang server code.

set -euo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PROJECT_DIR="$(dirname "$DIR")"

# See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

cd $PROJECT_DIR

export GOBIN=$PROJECT_DIR/bin
export PATH=$GOBIN:$PATH

echo "+ Install swagger tools"
go install github.com/swaggo/swag/cmd/swag

echo "+ Clean up go mod"
go mod tidy

echo "+ Generate swagger spec"
swag init -g cmd/chaos-dashboard/main.go
