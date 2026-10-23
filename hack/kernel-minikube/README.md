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

# Stock minikube capability evidence

The `minikube-capability` job in `kernel_image.yml` checks whether a GitHub-hosted Ubuntu runner can start one KVM-backed minikube node. It uses the stock ISO, minikube 1.38.1, Kubernetes v1.34.5, and containerd. The node receives 2 CPUs, 4096 MiB of RAM, and a 16 GiB disk.

The `kernel-image` job runs only on `workflow_dispatch` when the boolean input `run-qualification` is true. The input defaults to false. PR runs skip `kernel-image` and its dependent `image-runtime` and `lifecycle-prepare` jobs. Those skips provide no image or fault qualification coverage. `lifecycle-prepare` also requires a PR event, so it remains skipped on manual runs. The capability job runs independently on the workflow's existing PR trigger. It does not depend on workflow dispatch being available on the default branch.

`probe.py` accepts `host PROFILE OUTPUT`, `guest PROFILE OUTPUT`, and `cleanup PROFILE OUTPUT`. `PROFILE` must equal `cm-kvm-${GITHUB_RUN_ID}-${GITHUB_RUN_ATTEMPT}`. The host stage requires a GitHub-hosted Linux runner, virtualization CPU flags, a working KVM API, 2 CPUs, 5 GiB of available RAM, and 20 GiB of free disk. It probes KVM with sudo before the workflow installs libvirt and grants the runner access through ACLs.

The job makes one startup attempt with a 10-minute step limit. It has no restart or driver fallback. Default action caches remain shared. An unexpected setup or observation failure stops subsequent work except diagnostics and cleanup. Every observation command has a deadline. Cleanup deletes the exact profile once and checks that its profile directories and libvirt domain are absent. Cleanup also compares libvirt networks with the pre-start inventory and reports new networks left behind. It never prunes global resources or caches.

The artifact contains `result.json` and command evidence. The result separates these observations.

- `host` records `proven`, `failed`, or `untested` host capability.
- `transport` records `proven`, `failed`, or `untested` stock-node readiness and SSH observations.
- `fault_prerequisites` records `unsupported` when a required property is absent, `unknown` when evidence is unavailable, `observed` when all listed properties are visible, or `failed` when observation fails. `untested` means the guest stage did not run.
- `cleanup` records `removed`, `failed`, `not_created`, or `untested`. A skipped start records `not_created` without invoking minikube.

Guest observations read the kernel release, kernel configuration, `should_failslab` symbol, ERRNO allowlist eligibility, matching headers, and process namespace links. They do not mount debugfs, attach BPF programs, deploy candidate images, or create KernelChaos resources. An unreadable allowlist remains unknown. `fault_qualification` is always `untested`, including when every prerequisite is observed.

The existing candidate `bzImage` is not a drop-in replacement for the stock ISO kernel. Its modules and boot compatibility have not been qualified by this job.

The offline boundary checks run with `python3 -B -m unittest discover -s hack/kernel-minikube -p 'test_*.py'`. They do not start minikube or prove runner capability.

The capability job initializes its evidence directory from `RUNNER_TEMP` and exports `EVIDENCE` through `GITHUB_ENV`. Before host checks, actionlint v1.7.12 validates the workflow and writes `actionlint.log` to that directory. The same local check is:

```bash
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 -shellcheck= -pyflakes= .github/workflows/kernel_image.yml
```

ShellCheck and Pyflakes are disabled for this Actions expression check. Bash and Python checks remain separate.

The independent `kernel-workflow-validation` job in `ci.yml` reports actionlint errors even when `kernel_image.yml` cannot start.
