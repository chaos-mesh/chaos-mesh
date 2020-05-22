#!/usr/bin/env bash
#
# Generate all protobuf bindings.
# Run from repository root.
set -eu

PROTOC_BIN=${PROTOC_BIN:-protoc}

if ! [[ "$0" =~ "hack/genproto.sh" ]]; then
	echo "must be run from repository root"
	exit 255
fi

DIRS="pkg/chaosdaemon/pb pkg/chaosfs/pb"

echo "generating code"
for dir in ${DIRS}; do
	pushd ${dir}
		${PROTOC_BIN} --go_out=plugins=grpc:. -I=. *.proto
	popd
done
