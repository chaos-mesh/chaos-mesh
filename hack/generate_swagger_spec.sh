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

# This script call swag converts Go annotations to Swagger Documentation.
#
# See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

set -euo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PROJECT_DIR="$(dirname "$DIR")"

cd $PROJECT_DIR

export GOBIN=$PROJECT_DIR/bin
export PATH=$GOBIN:$PATH

echo "+ Install swaggo/swag"
go install github.com/swaggo/swag/cmd/swag

echo "+ Generate swagger spec"
swag init -g cmd/chaos-dashboard/main.go --output pkg/dashboard/swaggerdocs --pd --parseInternal

echo "+ Update go mod"
go mod tidy
