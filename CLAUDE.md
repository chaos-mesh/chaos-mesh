# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Chaos Mesh is a cloud-native Chaos Engineering platform for Kubernetes that provides various types of fault simulation. It consists of three main components:

1. **Chaos Controller Manager**: Core component responsible for scheduling and managing Chaos experiments through various CRD controllers (Workflow, Scheduler, and fault type controllers)
2. **Chaos Daemon**: Runs as a DaemonSet with privileged permissions, interferes with network devices, file systems, and kernels by accessing target Pod namespaces
3. **Chaos Dashboard**: Web UI for managing, designing, and monitoring Chaos experiments

## Development Commands

### Build Environment Setup
```bash
make image-build-env
make image-dev-env
make enter-devenv
make enter-buildenv
```

### Code Generation
```bash
make generate
make config
make chaos-build
make proto
make swagger_spec
make generate-deepcopy
make generate-client
make generate-clientset
make generate-lister
make generate-informer
make manifests/crd.yaml
```

### Quality Checks
```bash
make check
make fmt
make lint
make vet
make tidy
make gosec-scan
```

### Building
```bash
make all
make image
make ui
```

### Testing
```bash
make test
make coverage
make e2e
make e2e-build
make failpoint-enable
make failpoint-disable
```

### Local Development
Build specific components:
```bash
make local/chaos-daemon
make local/chaos-controller-manager
make local/chaos-dashboard
```

## Architecture

### Controller Design Principles
Controllers in Chaos Mesh follow strict design principles documented in `controllers/README.md`:

1. **One Controller Per Field**: Each field is controlled by at most one controller to avoid conflicts and hidden bugs
2. **Standalone Operation**: Controllers work independently without depending on other controllers
3. **Simple Behavior**: Controller logic should be describable in ~100 words
4. **Error Handling**: Use `ctrl.Result{Requeue: true}, nil` for retriable errors to leverage exponential backoff

### Chaos Types
Chaos implementations are in `controllers/chaosimpl/`:
- `awschaos`: AWS fault injection
- `azurechaos`: Azure fault injection
- `blockchaos`: Block device faults
- `dnschaos`: DNS fault injection
- `gcpchaos`: GCP fault injection
- `httpchaos`: HTTP fault injection
- `iochaos`: I/O fault injection
- `jvmchaos`: JVM fault injection
- `kernelchaos`: Kernel fault injection
- `networkchaos`: Network fault injection
- `physicalmachinechaos`: Physical machine faults
- `podchaos`: Pod lifecycle faults
- `stresschaos`: CPU/Memory stress
- `timechaos`: Time skew simulation

### API Structure
CRD definitions are in `api/v1alpha1/` with types, webhooks, and tests for each chaos kind.

### Key Directories
- `cmd/`: Main entry points for binaries (controller-manager, daemon, dashboard, builder)
- `controllers/`: Controller implementations and reconciliation logic
- `pkg/`: Shared packages (chaosdaemon, dashboard, grpc, selector, metrics, etc.)
- `api/`: CRD API definitions (v1alpha1)
- `config/`: Kubernetes manifests (CRD, RBAC, webhook)
- `helm/chaos-mesh/`: Helm chart for deployment
- `e2e-test/`: End-to-end test suite
- `images/`: Dockerfiles for all components
- `hack/`: Build and development scripts
- `ui/`: Frontend dashboard (pnpm-based)

## Development Workflow

1. **Before Making Changes**: Run `make check` to ensure code passes all checks
2. **After Code Changes**:
   - Run `make generate` if modifying CRDs or APIs
   - Run `make manifests/crd.yaml` to update CRD manifests
   - Run `make check` before committing
3. **Testing**: Run `make test` for unit tests
4. **Building Images**: Use `make image` to build all component images
5. **Commits**: Use `git commit --signoff` for DCO compliance

## Testing

- **Unit Tests**: `make test` runs tests with failpoint support and generates coverage
- **E2E Tests**: `make e2e` runs end-to-end tests in current Kubernetes cluster
- **Test Utilities**: Build with `make test-utils` (timer, multithread_tracee, fakeclock)
- **Coverage**: `make coverage` generates coverage reports

## Code Style

- Use `goimports` with `-local github.com/chaos-mesh/chaos-mesh` for formatting
- Run `revive` linter with configuration in `revive.toml`
- Keep `go.mod` tidy across all submodules (root, api, e2e-test)
- Use `controller-gen` for CRD and RBAC generation

## Build System

The project uses a containerized build environment with two Docker images:
- **build-env**: For compiling binaries (minimal build tools)
- **dev-env**: For development tasks (code generation, linting, testing)

Generated makefiles (`binary.generated.mk`, `container-image.generated.mk`) are created by `make generate-makefile`.

## UI Development

The dashboard UI is in `ui/` and uses pnpm:
```bash
cd ui
pnpm install --frozen-lockfile
pnpm build
```

Set `UI=1` environment variable to include UI in builds.

## Common Pitfalls

1. Always run `make check` before creating a PR
2. When modifying CRD structs, regenerate with `make generate && make manifests/crd.yaml`
3. Use failpoint carefully in tests (enable/disable properly)
4. Multi-module project: run `go mod tidy` in root, api/, and e2e-test/ directories
5. Controller modifications should follow "one controller per field" principle
