<!--
Copyright 2026 Chaos Mesh Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

# Kernel image guest qualification

The `Kernel image qualification` workflow builds and checks the installed image
without faults on native amd64 and arm64 runners. A separate amd64 job runs the
exact saved image's installed runtime on Linux 6.8 in one disposable guest.
It runs automatically for the workflow's path filters and can also be dispatched
manually. The guest adapter requires a non-root GitHub-hosted amd64 runner.

The image Dockerfile, workflow checkout and `manifest.py` pin bpfki to the same
immutable commit. The adapter builds the original-source control with bpfki's
three build-file shims and builds only the `kernel-tests` target from the pinned
candidate source. The candidate remains `/usr/local/bin/bpfki` from the image.
The original probe, comparison driver, kernel configuration, kernel builder,
guest init, launcher and result reader are reused unchanged.

The producer artifact records the image/config ID, layer identities, saved
archive checksum and archive manifest checksum. Packaging checks all six
installed executables, both installed program files, the dynamic loader and
resolved shared libraries, BCC and LLVM headers, and Clang builtin headers before and after
adding guest utilities and exporting the filesystem. Evidence includes the
source-built BCC tag/commit and CMake configuration, pinned source hashes,
kernel/header hashes, guest logs and comparison results. No registry publication
is needed.

The guest has 2 vCPUs and 3072 MiB RAM, no network interface, one owned root disk
and a read-only share of matching kernel headers. The unchanged harness allows
a 120-second polling allowance per case, an 8-minute deadline checked before
each case, 540 seconds for the launcher and
720 seconds for QEMU, with its existing termination grace. Each worker performs
at most 16 operations. The job limit is 60 minutes.

Rootfs assembly uses `mkfs.ext4 -d` without a host mount. It chooses a fixed 5 GiB
disk when the measured contents, filesystem overhead allowance and 1 GiB of
usable free space fit; otherwise it permits a fixed 8 GiB disk. It requires a
further 4 GiB of host free space and fails if these bounds cannot be met. Kernel
builds reuse `~/.cache/bpfki/kernel`; image additions reuse the Ubuntu 24.04 amd64
BuildKit apt cache. No modules are built, installed or loaded.

Passing requires all three candidate fault profiles, unaffected controls, two
recovery cycles in the same processes, SLAB invalid-update preservation, exact
original-source defect reproduction with `functional_pass=false`, cleanup and
complete matching image evidence. BIO must also retain the same owned
FTRACE probe before and after its finite operation, with complete descriptor and
address evidence and no OPTIMIZED, DISABLED or GONE flags. An unsupported kernel, missing evidence or a
timeout fails qualification. Failure logs are uploaded and the owned disk and
staging are removed.

A passing run qualifies the image's runtime filesystem on the controlled amd64 kernel.
Kubernetes/chart integration, ARM fault injection and other kernels require
separate qualification.
