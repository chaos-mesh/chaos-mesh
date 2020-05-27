#!/usr/bin/env bash

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
swag init -g cmd/chaos-server/main.go
