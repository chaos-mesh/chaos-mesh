#!/usr/bin/env bash
# Copyright 2026 Chaos Mesh Authors.
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

image=${1:?usage: check-kernel-image.sh IMAGE PLATFORM}
platform=${2:?usage: check-kernel-image.sh IMAGE PLATFORM}
state_dir=$(mktemp -d)
cleanup() {
  status=$?
  trap - EXIT
  if [[ -s "$state_dir/container-id" ]]; then
    docker rm -f "$(cat "$state_dir/container-id")" >/dev/null 2>&1 || true
  fi
  rm -rf "$state_dir"
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

docker run --rm -i --cidfile "$state_dir/container-id" --platform "$platform" \
  --network none --cap-drop ALL --security-opt no-new-privileges \
  --read-only --user 65534:65534 --pids-limit 64 --memory 1g \
  --tmpfs /tmp:rw,exec,nosuid,nodev,size=512m "$image" \
  timeout 180s bash -seuo pipefail <<'CHECK'
    for binary in bpfki time-chaos fail-bio fail-mount fail-ext4-mount fail-alloc-page; do
      test -x "/usr/local/bin/$binary"
      ldd "/usr/local/bin/$binary" > /tmp/libraries
      cat /tmp/libraries
      if grep -q "not found" /tmp/libraries; then exit 1; fi
    done
    for program in clock_gettime gettimeofday; do
      cmp "/bpfki/progs/$program" "/usr/local/bin/$program"
    done
    cmake -S /bpfki -B /tmp/smoke -DBPFKI_BUILD_SMOKE_TEST=ON
    cmake --build /tmp/smoke --target bpfki-smoke --parallel 2
    cd /usr/local/bin
    ./bpfki --host=127.0.0.1 --port=50051 > /tmp/service.log 2>&1 &
    service_pid=$!
    trap 'kill "$service_pid" 2>/dev/null || true; wait "$service_pid" 2>/dev/null || true; cat /tmp/service.log' EXIT
    /tmp/smoke/tests/smoke/bpfki-smoke 127.0.0.1:50051
    kill -0 "$service_pid"
    echo "PASS: installed kernel image files, dynamic libraries and no-fault RPC communication."
CHECK
