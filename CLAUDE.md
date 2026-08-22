# Chaos Mesh repository guide for AI coding agents

## Project overview

Chaos Mesh is a cloud-native chaos engineering platform for Kubernetes. Its main runtime components are:

1. **Chaos Controller Manager**: reconciles Chaos Mesh CRDs, including chaos experiments, schedules, workflows, status checks, and multi-cluster resources.
2. **Chaos Daemon**: runs as a privileged DaemonSet and performs node-level fault injection through container runtimes, network devices, filesystems, processes, and kernel facilities.
3. **Chaos Dashboard**: a Go API server plus a React web UI for creating and observing experiments.

## Repository map

- `api/v1alpha1/`: CRD types, validation/defaulting webhooks, and generated API helpers. `api/` is a separate Go module.
- `controllers/`: controller-runtime reconcilers. `chaosimpl/` contains the injection/recovery implementations; `common/`, `action/`, `schedule/`, `statuscheck/`, and `multicluster/` provide the surrounding control logic. Pod-level HTTP, I/O, and network resources have dedicated controllers under `podhttpchaos/`, `podiochaos/`, and `podnetworkchaos/`.
- `pkg/`: shared runtime packages, including daemon services, dashboard APIs and persistence, workflow controllers, selectors, metrics, and generated clients.
- `cmd/`: binary entry points and repository generators.
- `config/`, `helm/chaos-mesh/`, and `manifests/`: Kubernetes, Helm, and generated deployment artifacts.
- `e2e-test/`: Ginkgo/Godog end-to-end suites and the e2e helper. It contains two additional Go modules: `e2e-test/` and `e2e-test/cmd/e2e_helper/`.
- `ui/`: pnpm workspace containing the React/TypeScript/Vite dashboard and the OpenAPI code generator.
- `build/`, `hack/`, and `images/`: containerized build tooling, helper scripts, and image definitions.

The repository currently has four Go modules: the root module, `api/`, `e2e-test/`, and `e2e-test/cmd/e2e_helper/`. All declare Go 1.25.11. CI uses Node.js 24 and pnpm 11 for the UI. Treat the checked-in `go.mod`, UI package manifests, and CI workflows as the source of truth if these versions change.

## Working safely in this repository

- Inspect `git status --short` before editing and preserve unrelated user changes. Do not discard generated or formatted changes without establishing where they came from.
- Prefer the narrowest relevant test while iterating. Run the full repository checks only after the focused checks pass and when their Docker/time cost is appropriate for the task.
- Most top-level Make targets use the containerized dev or build environment and may build Docker images on first use. Host-only binary targets are explicitly prefixed with `local/`.
- `make check`, `make test`, and `make generate` are not read-only: they can rewrite generated code, manifests, module files, formatted Go files, or test artifacts. Review the resulting diff and keep only task-related changes.
- Do not run e2e tests unless the task requires them and a disposable Kubernetes cluster is available. The suite installs resources and injects real faults.
- Do not edit generated files directly. Change their source and run the matching generator.

## Development commands

Run `make help` to see the targets supported by the current checkout. Makefile evaluation computes container image tags, so even `make help` can emit a Docker socket warning when Docker is unavailable; the target list is still usable.

### Build environments

```bash
make image-dev-env
make image-build-env
make enter-devenv
make enter-buildenv
```

The dev environment contains code-generation, lint, and test tools. The build environment is used for production binaries and native helper objects.

### Focused Go tests

Use normal Go package selection for quick iteration when the host has the required Go/CGO and envtest dependencies:

```bash
go test ./path/to/package
go test ./path/to/package -run TestName
(cd api && go test ./v1alpha1/...)
```

Run tests from the module that owns the package. The canonical repository-wide test is:

```bash
make test
```

`make test` first runs generation and builds native test utilities, enables failpoint instrumentation, runs root and `api/` package tests serially with coverage, and then disables failpoints. If it is interrupted or fails before cleanup, run `make failpoint-disable` before continuing. `make coverage` renders reports from an existing `cover.out`; run `make test` first when that file is not current.

### Verification

During normal development, do not default to `make check` after every edit. Use focused package tests, builds, and the change-specific checks below while iterating. Run only the relevant standalone check when the change has a limited scope.

Reserve the aggregate check for final validation, broad cross-cutting changes, or when explicitly requested:

```bash
make check
```

`make check` is the core of the main CI verification job. It runs repository-wide generation, `go vet`, `revive`, `goimports`, `go mod tidy` across all four modules, and Helm values schema generation. Some steps rewrite tracked files, so inspect `git diff` afterward. CI follows it with `git diff --quiet` and requires those checks to leave no uncommitted generated or formatted output.

`make fmt` formats all hand-written Go files with `goimports`, using `-local github.com/chaos-mesh/chaos-mesh`; it is not a focused, single-package command. The standalone `make vet`, `make lint`, and `make tidy` targets run the corresponding portions of `make check`.

`make gosec-scan` is a separate, non-blocking security report and is not part of `make check`.

### Code generation

```bash
make generate
make proto
make generate-makefile
```

- `make generate` is the umbrella generator for CRD/Helm manifests, deepcopy methods, typed clients/informers/listers, Chaos Mesh API helpers, and dashboard Swagger output. It already includes `manifests/crd.yaml`; do not run that target again unless only the combined manifest is needed.
- Run `make proto` after changing daemon or kernel protobuf definitions.
- The root `binary.generated.mk`, `local-binary.generated.mk`, and `container-image.generated.mk` files say `DO NOT EDIT`. Change the generator under `cmd/generate-makefile/` and run `make generate-makefile` instead.
- Files named `zz_generated.*`, `pkg/client/**`, protobuf outputs, Swagger docs, and UI files carrying an auto-generated header must be regenerated from their source rather than patched by hand.

After an API/CRD change, inspect all generated Go, `config/crd/bases/`, `helm/chaos-mesh/crds/`, and `manifests/crd.yaml` changes together.

### Building binaries and images

Host builds are available for the two targets declared in `local-binary.generated.mk`:

```bash
make local/chaos-controller-manager
make local/chaos-dashboard
```

There is no `local/chaos-daemon` target. The daemon needs CGO/native helpers and is built through its generated build-environment target or as an image.

```bash
make image
make all
```

`make image` builds the controller-manager, daemon, and dashboard container images. `make all` also generates the combined CRD manifest. Both are expensive Docker builds and are not substitutes for focused compilation or tests.

### UI development

The current UI uses React, TypeScript, Vite, Material UI, TanStack Query, and Zustand. Run UI commands from `ui/`:

```bash
pnpm install --frozen-lockfile
pnpm build
pnpm test
pnpm -F @ui/app lint
VITE_API_BASE_URL=http://localhost:2333 pnpm start
```

The Vite development server listens on port 3000 and proxies `/api` to `VITE_API_BASE_URL`. For a non-interactive/CI install, set `HUSKY=0` if Git hook installation is unwanted.

The top-level Make target only builds and embeds UI assets when `UI=1` is set:

```bash
UI=1 make ui
```

## Controller design rules

Follow `controllers/README.md` when changing reconcilers:

1. A field must be owned by at most one controller. Do not introduce competing writers for the same desired state.
2. A controller must have understandable standalone behavior and must not rely on undocumented ordering between independent controllers.
3. Keep controller behavior small and document it; if it cannot be summarized clearly, reconsider the resource or controller boundary.
4. For a retriable condition that should use the workqueue's rate limiter, the repository convention is `ctrl.Result{Requeue: true}, nil`. Preserve the surrounding controller's established error semantics rather than applying this mechanically to every error.

Keep API validation/defaulting in `api/v1alpha1/`, orchestration and status transitions in controllers, and privileged node operations behind the Chaos Daemon APIs. Reuse the common action/pipeline, selector, recorder, and daemon client abstractions instead of bypassing them from an individual chaos type.

## Change-specific checks

- **Go implementation:** run focused package tests and formatting; then use `make check`/`make test` when full validation is warranted.
- **CRD or webhook:** run `make generate`, focused `api/` tests, and inspect all manifests and generated clients.
- **Protobuf:** run `make proto` and test both the client and server packages.
- **Dashboard Go API:** run focused `pkg/dashboard/...` tests and regenerate Swagger/OpenAPI-dependent artifacts when the API surface changes.
- **UI:** run `pnpm -F @ui/app lint`, `pnpm build`, and `pnpm test` from `ui/`.
- **Helm values/chart:** regenerate `helm/chaos-mesh/values.schema.json` with `make helm-values-schema` and run relevant `pkg/helm/...` tests.
- **Generated build inventory:** change `cmd/generate-makefile/`, run `make generate-makefile`, and inspect all three generated Makefiles.

If creating commits, use `git commit --signoff` for DCO compliance. Do not create commits unless the task explicitly asks for them.
