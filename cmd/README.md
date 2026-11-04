# Command entry points

The `cmd/` directory contains executable entry points and repository development tools. Each child directory builds one command. Most application behavior should live in reusable packages under `controllers/`, `pkg/`, or `api/`; packages under `cmd/` should mainly parse configuration, assemble dependencies, manage process lifecycle, and delegate to those packages.

## Directory guide

| Directory | Purpose | Primary build path |
| --- | --- | --- |
| `chaos-controller-manager/` | Starts the Kubernetes controller manager, admission webhooks, metrics, and controller modules through Uber Fx. | `make local/chaos-controller-manager` or `make image-chaos-mesh` |
| `chaos-daemon/` | Starts the privileged per-node gRPC/HTTP daemon that performs fault injection. | `make images/chaos-daemon/bin/chaos-daemon` or `make image-chaos-daemon` |
| `chaos-daemon-helper/` | Builds `cdh`, a small Cobra command executed by Chaos Daemon inside another namespace or cgroup. | `make images/chaos-daemon/bin/cdh`; included in `make image-chaos-daemon` |
| `chaos-dashboard/` | Starts the dashboard API server, persistence layer, collectors, and TTL controllers through Uber Fx. Its `main.go` also contains the root Swagger annotations. | `make local/chaos-dashboard` or `make image-chaos-dashboard` |
| `watchmaker/` | Builds a Linux-specific helper for process time skew. The Darwin entry point only reports that the platform is unsupported. | `make watchmaker` |
| `chaos-builder/` | Scans markers on `api/v1alpha1` types and generates API helpers, tests, workflow/schedule registration, and a frontend type map. | `make chaos-build`; normally invoked through `make generate` |
| `generate-makefile/` | Generates the three root Makefiles for local binaries, product binaries, and container images from checked-in recipe definitions. | `make generate-makefile` |

The binaries shipped in product images are `chaos-controller-manager`, `chaos-daemon`, `cdh`, and `chaos-dashboard`. `watchmaker` is a specialized runtime helper. `chaos-builder` and `generate-makefile` are repository development tools and are not deployed as Chaos Mesh services.

## Keep entry points small

An entry point should contain only the code that is specific to starting and stopping the process:

- command-line flag or environment configuration parsing;
- logger, metrics, and dependency-injection setup;
- construction of the top-level server, manager, or application;
- signal handling, startup, and graceful shutdown;
- version reporting and process exit behavior.

Put controllers, API handlers, daemon operations, storage logic, and other testable behavior in an importable package. This keeps the implementation reusable and allows tests to exercise it without starting a real process. Existing entry points delegate to these main packages:

- controller wiring and reconciler modules: `cmd/chaos-controller-manager/provider/` and `controllers/`;
- daemon server and helper commands: `pkg/chaosdaemon/`;
- dashboard configuration and services: `pkg/config/` and `pkg/dashboard/`;
- time manipulation: `pkg/time/`.

Do not import one `package main` from another package. Move shared code into an appropriately scoped library package instead.

### Build-rule caveat

The generated product-binary recipes currently compile an explicit `cmd/<name>/main.go` path, not the whole package directory. A sibling `.go` file in one of those command directories is therefore not automatically included in the generated local or image build.

Prefer keeping the entry point in `main.go` and moving additional logic into an importable package. If an entry point genuinely needs multiple source files, update its `SourcePath` in `cmd/generate-makefile/`, regenerate the Makefiles, and verify every affected local and image build target.

## Configuration and dependency wiring

Avoid duplicating configuration definitions in an entry point. Use the existing source of truth for each component:

- Controller Manager environment configuration is defined in `pkg/config/controller.go`, loaded through `controllers/config/`, and consumed by the Fx modules in `cmd/chaos-controller-manager/provider/` and `controllers/`.
- Chaos Daemon command-line flags populate `pkg/chaosdaemon.Config` in `cmd/chaos-daemon/main.go`; server behavior belongs in `pkg/chaosdaemon/`.
- Dashboard environment configuration is defined in `pkg/config/dashboard.go`; API, storage, collector, and TTL behavior belongs under `pkg/dashboard/`.
- `cdh` subcommands are defined under `pkg/chaosdaemon/helper/` and registered by `cmd/chaos-daemon-helper/main.go`.

When adding configuration, update the owning config type, its tests, deployment manifests or Helm values, and documentation together. Preserve the existing Fx module boundaries for Controller Manager and Dashboard instead of constructing their internal dependencies directly in `main.go`.

## Developing an existing command

1. Trace the entry point to the package that owns the behavior and make the change there when possible.
2. Add focused unit tests in the owning package. Keep tests under `cmd/` for entry-point wiring or command-generator behavior.
3. Run the narrowest relevant tests while iterating.
4. Build the affected binary through its supported target before requesting review.
5. Run broader repository checks only when the change warrants them or as final validation.

Useful focused commands include:

```bash
go test ./cmd/chaos-controller-manager/provider/...
go test ./pkg/chaosdaemon/...
go test ./pkg/dashboard/...
go test ./cmd/chaos-builder/...

make local/chaos-controller-manager
make local/chaos-dashboard
make images/chaos-daemon/bin/chaos-daemon
make images/chaos-daemon/bin/cdh
```

Chaos Daemon and the Linux implementation of watchmaker depend on Linux facilities, CGO or native helper objects, and—in normal operation—elevated privileges. Validate those changes in a Linux environment and use their Make targets so the required native dependencies are built. Do not execute fault-injection helpers directly on a development machine unless that is explicitly part of a controlled test.

## Changing generated code

### Chaos API generation

`chaos-builder` scans `api/v1alpha1` for `+chaos-mesh:*` markers and writes:

- `api/v1alpha1/zz_generated.chaosmesh.go` and its generated tests;
- workflow and schedule registration files under `api/v1alpha1/`;
- `ui/app/src/api/zz_generated.frontend.chaos-mesh.ts`.

Run it from the repository root through Make so its relative input and output paths resolve correctly:

```bash
make chaos-build
```

For normal CRD changes, prefer `make generate`, which also updates deepcopy code, clients, CRD manifests, and dashboard Swagger output. Never patch the generated outputs directly; change the API type, marker, or generator template and then inspect the complete generated diff.

### Build Makefile generation

Do not edit these files directly:

- `binary.generated.mk`;
- `local-binary.generated.mk`;
- `container-image.generated.mk`.

Their templates and recipe lists live in `cmd/generate-makefile/`. After changing a binary or image recipe, run:

```bash
go test ./cmd/generate-makefile/...
make generate-makefile
git diff -- binary.generated.mk local-binary.generated.mk container-image.generated.mk
```

Inspect all three outputs even when only one recipe was intentionally changed.

### Dashboard API documentation

The root Swagger metadata is attached to `cmd/chaos-dashboard/main.go`, while endpoint annotations live with the dashboard handlers. After changing the documented API surface, run:

```bash
make swagger_spec
```

Review the generated files under `pkg/dashboard/swaggerdocs/` together with the API change.

## Adding a new command

Before adding a binary, confirm that it needs an independent process rather than a package, controller, daemon RPC, dashboard endpoint, or subcommand of an existing executable. If a new command is appropriate:

1. Create `cmd/<command>/main.go` with the project copyright and Apache 2.0 header.
2. Keep `main` focused on configuration, wiring, lifecycle, and error reporting; implement behavior in an importable package.
3. Add focused unit tests for the implementation and any nontrivial command wiring.
4. Add the binary to `binaryRecipes` in `cmd/generate-makefile/generate_binary_makefile.go` when it is part of a product build.
5. Add it to `localBinaryRecipes` only when a host build is supported and useful.
6. If it is shipped in a container, update `containerImageRecipes`, the relevant `images/<image>/Dockerfile`, and the image dependency targets.
7. Run `make generate-makefile` and inspect all generated Makefiles.
8. Update Helm deployments, RBAC, services, image configuration, and release packaging when the process is deployed in Kubernetes.
9. Add `--version` behavior with `pkg/version` for user-facing binaries, following the existing commands.
10. Verify the focused tests and exact build target, then perform the appropriate final repository checks.

If the new command is a repository-only tool, a direct Make target may be clearer than adding it to product binary or image recipes.

## Review checklist

- Is `main.go` limited to process concerns rather than application logic?
- Are configuration and dependency wiring consistent with the owning component?
- Are Linux-only or privileged operations isolated and tested safely?
- Are generated files changed through their generator?
- Do focused tests cover the behavior that changed?
- Does the supported local, build-environment, or image target compile the command?
- If a binary or image was added, were generated Makefiles and deployment assets updated?
