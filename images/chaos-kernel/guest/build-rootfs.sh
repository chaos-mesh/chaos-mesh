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

if [[ $# != 4 || ${GITHUB_ACTIONS:-} != true || ${RUNNER_ENVIRONMENT:-} != github-hosted ||
      ${RUNNER_OS:-} != Linux || ${RUNNER_ARCH:-} != X64 ||
      ${RUNNER_TEMP:-} != /home/runner/work/_temp || $EUID == 0 ||
      $(uname -s) != Linux || $(uname -m) != x86_64 ]]; then
  echo 'Usage on a disposable GitHub-hosted amd64 runner: build-rootfs.sh BPFKI_SOURCE KERNEL_TREE IMAGE_ARTIFACT OUTPUT' >&2
  exit 2
fi
source_dir=$(realpath "$1")
kernel=$(realpath "$2")
artifact=$(realpath "$3")
mkdir -p "$4"
output=$(realpath "$4")
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
pin=f75f82881950d1f66f5b479e387f1c1f6fc6e9a7
test "$(git -C "$source_dir" rev-parse HEAD)" = "$pin"
git -C "$source_dir" diff --exit-code HEAD
git -C "$source_dir" cat-file -e 2825d3832c9230ab6e26e8a47ed6de8d16ea5ae3^{commit}

context=$(mktemp -d "$RUNNER_TEMP/chaos-kernel-rootfs-XXXXXXXX")
container=
fixture="chaos-kernel-guest:$GITHUB_RUN_ID-$GITHUB_RUN_ATTEMPT"
base="chaos-kernel-guest-base:$GITHUB_RUN_ID-$GITHUB_RUN_ATTEMPT"
cleanup() {
  status=$?
  trap - EXIT
  [[ -z $container ]] || docker rm -f "$container" >/dev/null 2>&1 || true
  docker image rm "$fixture" >/dev/null 2>&1 || true
  docker image rm "$base" >/dev/null 2>&1 || true
  sudo rm -rf -- "$context"
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

python3 "$script_dir/manifest.py" image "$artifact/image.tar.gz" "$artifact/image-inspect.json" > "$output/image.json"
cmp "$output/image.json" "$artifact/image.json"
image_id=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["image_id"])' "$output/image.json")
docker load -i "$artifact/image.tar.gz"
docker image inspect "$image_id" > "$output/loaded-image-inspect.json"
python3 "$script_dir/manifest.py" image "$artifact/image.tar.gz" "$output/loaded-image-inspect.json" > "$context/loaded-image.json"
cmp "$output/image.json" "$context/loaded-image.json"
docker tag "$image_id" "$base"
test "$(docker image inspect --format '{{.Id}}' "$base")" = "$image_id"
rm "$artifact/image.tar.gz"
docker run --rm -i --network none --cap-drop ALL --security-opt no-new-privileges \
  --read-only --pids-limit 64 --memory 1g --entrypoint python3 "$image_id" \
  - capture < "$script_dir/manifest.py" > "$output/runtime.json"

(cd "$source_dir" && python3 tests/kernel/prepare-control.py)
mkdir "$context/candidate"
git -C "$source_dir" archive "$pin" CMakeLists.txt src proto progs examples tests | tar -xf - -C "$context/candidate"
chmod 755 "$context/candidate/tests/kernel/vm/init.sh"
python3 "$script_dir/manifest.py" source "$context/candidate" > "$output/source-files.json"
cp -a "$RUNNER_TEMP/bpfki-source-control-src" "$context/source-control"
cp "$RUNNER_TEMP/bpfki-source-control.json" "$context/source-control.json"
cp "$script_dir/manifest.py" "$context/manifest.py"
cp "$output/runtime.json" "$context/runtime.json"
docker build --pull=false --progress=plain --build-arg "KERNEL_IMAGE=$base" \
  -f "$script_dir/Dockerfile" -t "$fixture" "$context"
test "$(docker image inspect --format '{{.Id}}' "$base")" = "$image_id"
docker image inspect "$fixture" > "$output/guest-image-inspect.json"
python3 - "$output/image.json" "$output/guest-image-inspect.json" <<'PYTHON'
import json, sys
base = json.load(open(sys.argv[1]))
guest = json.load(open(sys.argv[2]))[0]
assert guest["Architecture"] == "amd64" and guest["Os"] == "linux"
assert guest["RootFS"]["Layers"][:len(base["layers"])] == base["layers"], "guest did not derive from the producer image layers"
PYTHON
container=$(docker create "$fixture")
mkdir "$context/rootfs"
docker export "$container" | sudo tar -xpf - -C "$context/rootfs"
docker rm "$container" >/dev/null
container=
release=$(cat "$kernel/bpfki-kernel-release")
sudo cp "$kernel/.config" "$context/rootfs/boot/config-$release"
sudo mkdir -p "$context/rootfs/lib/modules/$release"
sudo ln -s /kernel-headers "$context/rootfs/lib/modules/$release/build"
sudo ln -s /kernel-headers "$context/rootfs/lib/modules/$release/source"
python3 - "$context/ci.env" <<'PYTHON'
import os, shlex, sys
keys = ["GITHUB_ACTIONS", "RUNNER_ENVIRONMENT", "RUNNER_OS", "RUNNER_ARCH",
        "RUNNER_NAME", "RUNNER_TEMP", "GITHUB_REPOSITORY", "GITHUB_SHA",
        "GITHUB_RUN_ID", "GITHUB_RUN_ATTEMPT"]
with open(sys.argv[1], "w") as target:
    for key in keys:
        target.write(f"export {key}={shlex.quote(os.environ[key])}\n")
PYTHON
sudo cp "$context/ci.env" "$context/rootfs/etc/bpfki-ci.env"
sudo python3 "$script_dir/manifest.py" bundle "$context/rootfs" "$kernel" "$output"
sudo chown "$(id -u):$(id -g)" "$output/image-provenance.json"
sudo cp "$output/image-provenance.json" "$context/rootfs/home/runner/work/_temp/image-provenance.json"
sudo du -sx -B1 "$context/rootfs" > "$output/rootfs-allocated.txt"
sudo du -sx -B1 --apparent-size "$context/rootfs" > "$output/rootfs-apparent.txt"
python3 - "$output" <<'PYTHON'
import json, shutil, sys
from pathlib import Path
output = Path(sys.argv[1])
gib = 1024 ** 3
contents = max(int((output / name).read_text().split()[0]) for name in ["rootfs-allocated.txt", "rootfs-apparent.txt"])
free = shutil.disk_usage(output).free
size = next((size for size in [5, 8] if contents + size * gib // 16 + 128 * 1024 ** 2 + gib <= size * gib), None)
evidence = {"contents_bytes": contents, "filesystem_overhead_allowance": "disk/16 + 128 MiB",
            "minimum_guest_free_bytes": gib, "host_free_bytes": free, "host_reserve_bytes": 4 * gib,
            "disk_gib": size}
(output / "capacity.json").write_text(json.dumps(evidence, indent=2) + "\n")
assert size is not None, "rootfs exceeds the fixed 8 GiB capacity"
assert free >= (size + 4) * gib, "insufficient host storage for root disk plus 4 GiB reserve"
(output / "disk-size-gib.txt").write_text(str(size) + "\n")
PYTHON
truncate -s "$(cat "$output/disk-size-gib.txt")G" "$output/rootfs.ext4"
sudo mkfs.ext4 -q -F -d "$context/rootfs" "$output/rootfs.ext4"
sudo dumpe2fs -h "$output/rootfs.ext4" > "$output/rootfs-superblock.txt"
python3 - "$output/rootfs-superblock.txt" <<'PYTHON'
import re, sys
text = open(sys.argv[1]).read()
free = int(re.search(r"^Free blocks:\s+(\d+)", text, re.M)[1])
reserved = int(re.search(r"^Reserved block count:\s+(\d+)", text, re.M)[1])
size = int(re.search(r"^Block size:\s+(\d+)", text, re.M)[1])
assert (free - reserved) * size >= 1024 ** 3, "formatted rootfs has less than 1 GiB usable headroom"
PYTHON
