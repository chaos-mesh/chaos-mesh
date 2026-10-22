<!--
Copyright 2026 Chaos Mesh Authors.
Licensed under the Apache License, Version 2.0.
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an AS IS BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
-->

# Prepare the fixed KernelChaos lifecycle guest

The kernel-image workflow prepares this bundle after its actual-image runtime gate passes. Preparation does not boot a guest. The bundle requires review before a separate workflow change can activate the fixed launcher.

```sh
bash images/chaos-kernel/guest/lifecycle/prepare.sh \
  BPFKI_CHECKOUT IMAGE_DIR RUNTIME_EVIDENCE_DIR OUTPUT_DIR
```

`CM_LIFECYCLE_PROVENANCE` names the workflow's JSON provenance record. Preparation requires the same run's producer archive and complete passing runtime evidence. It retains the qualified kernel archive unchanged. Its original tag is `chaos-kernel:ci`; containerd resolves the chart reference as `docker.io/library/chaos-kernel:ci`. These references share the verified config and layers. Local images have no invented registry digest.

The output contains `bundle.tar.gz`, `bundle.sha256`, `preparation.json`, and `review/`. The archive contains `bundle.json`, one pristine sparse 16 GiB ext4 disk, `bzImage`, and finite header, payload, and review trees. Its inventory hashes every other launch input, including file modes and link targets. Extract with GNU tar's sparse-file support after checking the archive checksum and member paths. The manifest excludes its own recursive hash.

`inputs.lock.json` pins Ubuntu Base, Linux, release tools, and seven Kubernetes amd64 child manifests. Preparation checks downloaded bytes. Release-tool checksums do not claim signature verification. Ubuntu packages resolve per preparation run against the configured Ubuntu repositories. The sealed record includes signed InRelease files, verified signing fingerprints, package-index hashes, downloaded package hashes and origins, the installed package manifest, actual runtime-library resolution, and the installed module-loading audit.

Review the final kernel configuration, generated node and service configuration, image relationships, package and filesystem seals, and capacity report. Helm generates fresh certificates inside the guest. Review renders omit Secret objects and cannot predict those private bytes. No package resolution or guest input repair occurs during launch.

The separate CM kernel tree uses the shared `~/.cache/bpfki/kernel/linux-6.8.tar.xz` and a configuration/toolchain key under `~/.cache/bpfki/kernel/cm-linux-6.8/`. It leaves the direct-test kernel tree untouched. The build uses two jobs, builds only bzImage and prepared headers, and installs no kernel modules. Release assets and registry blobs use stable caches under `~/.cache/chaos-mesh/lifecycle/`. Ubuntu builds reuse `chaos-mesh-ubuntu24.04-apt-amd64`; normal CM builds retain the project's `GO_BUILD_CACHE`.

The inactive entry point is `run.sh EXTRACTED_BUNDLE OUTPUT_DIR`. A future consumer must verify the reviewed artifact and prior run through GitHub before invoking it from the exact prepared checkout. The launcher checks every sealed file and current host capacity, then starts one QEMU process with 2 CPUs, 6144 MiB RAM, the 16 GiB disk, no NIC, and only two read-only 9p shares. TCG is fixed. The guest deadline is boot age 840 seconds; the independent outer deadline is 900 seconds plus 20 seconds for termination. No retry is built in.

Launch requires at least 8 GiB MemAvailable and free host storage for the disk's remaining full allocation, a 4 GiB reserve, and evidence growth. It rejects concurrent builds. Serial bytes plus decoded required evidence must fit 8 MiB. Missing or truncated records fail. Both lifecycle cases, cleanup, actual systemd service exit, explicit guest exit, and QEMU exit must agree. Runtime changes to the disk are expected. The launcher preserves it for the workflow's owned-state cleanup and never describes the used disk as pristine.
