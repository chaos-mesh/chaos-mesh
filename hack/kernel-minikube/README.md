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

# Stock minikube KernelChaos qualification

`kernel_image.yml` runs the full qualification on matching pull requests and manual dispatches. Both `kernel-image` matrix entries build and check the candidate image on their native amd64 or arm64 runner. The amd64 artifact then feeds two independent paths.

- `image-runtime` retains the direct actual-image test of all three faults in its existing isolated guest.
- `component-images` builds controller-manager and chaos-daemon from the same checkout with the existing Make targets. It compiles both native fixture guard modes in the exact candidate kernel image and exports the fixed worker image and pinned CLI assets. `minikube-lifecycle` consumes these images and runs the normal chart on one stock KVM node.

The `qualification` job requires all four producer and runtime jobs to succeed. Skipped, failed, or cancelled prerequisites do not pass. The prior sealed `lifecycle-prepare` job is no longer part of CI. Its local scripts remain available and share the fixed lifecycle case engine with minikube.

## VM ownership

The workflow's existing setup action starts minikube 1.38.1 with Kubernetes v1.34.5, containerd, and the stock Linux 6.6.95 kernel. The node receives 2 CPUs, 6144 MiB of RAM, and a 16 GiB disk. There is one startup attempt and no kernel, ISO, or driver fallback.

`probe.py` binds `cm-kvm-${GITHUB_RUN_ID}-${GITHUB_RUN_ATTEMPT}` to its actual libvirt UUID, guest DMI UUID, boot ID, machine ID, node UID, and initial PID namespace. Host commands use the job's dedicated kubeconfig and explicit context. Native prepare and observation require the compiled minikube guard to match the live guest identity. The sealed fixture retains its separate compile-time guard.

`qualify.py` verifies the producer run, attempt, source revisions, archive hashes, image configs, ordered layer digests, and tool hashes before import. Guest observation runs in an unchanged candidate chaos-kernel image with hostPID inside the VM. Actual CRI identities bind each selected container to its image, Pod UID, PID, start time, cgroup, and held pidfd.

## Qualification and fixed cases

Before creating KernelChaos, a bounded host `kubectl exec` runs `qualify_bpf.py` in the chart's actual bpfki container. It checks the built-in root-owned header archive and BCC's automatic extraction. It loads and attaches an entry probe whose override helper is guarded by a fresh, empty, unpinned hash map. The script never inserts keys. Loaded bytecode must retain the helper. Map emptiness, entry attachment, probe detach, and owned FD release must all succeed.

The qualifier reads the allowlist through the bound guest PID 1 root after checking its boot ID, initial PID namespace, and debugfs mount. If BCC has no usable tracefs mount, it mounts tracefs in the actual sidecar only after verifying a separate mount namespace, a sysfs parent without shared propagation, and the guest tracefs superblock's visible `rw` options. A normal mount may normalize unobservable tracefs superblock metadata inside the disposable guest. Preparation writes no tracing controls. The mount remains through both cases and their recovery, and the owned VM teardown removes it. The chart and candidate images are unchanged.

`FAILSLAB` and `FAULT_INJECTION` are recorded as context. This SLAB override path instead requires the BPF override configuration, `should_failslab` entry, and its ERRNO allowlist entry. Passive prerequisites alone do not qualify a fault.

Strict-field and invalid-duration server dry-runs verify the normal admission path before the observer can arm. Failed or incomplete qualification prevents KernelChaos creation. The shared `cases.py` then runs expiration followed by UID-precondition deletion with these fixed limits.

- The active VM lease is 300 seconds. The two cases have 150-second and 90-second ceilings.
- Expiration lasts 120 seconds. First injection proof has at most 90 seconds.
- Each native operation has a three-second limit. The run performs eight target and eight control operations, including exactly four target ENOMEM operations.
- The final expiration fault must finish in the last seven seconds before expiry.
- UID Delete must be accepted within five seconds of the last deletion-case fault.
- Both cases require owned map removal and success from the original target and control processes within 30 seconds of the recovery trigger.

API traffic uses bounded standard-library HTTPS inside the observer. Absolute deadlines cover headers, partial JSON, and watch traffic. Mutations and worker operations are not retried after ambiguous results.

## Containment and evidence

An independent host watchdog starts before the no-fault BPF gate. It acknowledges a single active lease after setting its deadline. At expiry it stops the exact recorded libvirt UUID before profile deletion. The coordinator also stops that VM on ordinary failure. The workflow's `always()` cleanup removes the profile and checks the domain, profile directories, and added libvirt networks. Absence is safe to check again. No global resources or shared caches are pruned.

The artifact `kernel-minikube-lifecycle-ATTEMPT` includes the gate receipt, native and Kubernetes events, recovery result, containment result, cleanup result, and bounded command logs. Recovery and VM cleanup remain separate claims. Destroying a VM does not prove recovery. A successful lifecycle job requires both recovery cases and confirmed cleanup.

## Offline checks

Run these checks from the repository root.

```bash
python3 -B -m unittest discover -s hack/kernel-minikube -p 'test_*.py'
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 -shellcheck= -pyflakes= .github/workflows/kernel_image.yml .github/workflows/ci.yml
```

The unit tests exercise identity rejection, failed-gate arming, absolute HTTP deadlines, UID deletion, operation limits, and exact-UUID watchdog cleanup. They do not attach BPF or prove stock-kernel compatibility. Native compilation runs in `component-images`; actual header qualification and fault recovery require the CI runtime jobs.

Project Go caches, Docker build caches, and the pinned release-asset cache under `~/.cache/chaos-mesh/lifecycle/assets` are reused. The BCC extraction directory inside the disposable VM is runtime state.
