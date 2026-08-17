# Controller architecture and development guide

This directory contains the controller-runtime reconcilers used by Chaos Mesh. The controller manager assembles them through Uber Fx in `controllers.Module`, which is loaded by `cmd/chaos-controller-manager`.

Controllers should remain level-based: each reconciliation reads the current state, moves it toward the desired state, and can safely run again. A reconcile request identifies an object by namespace and name; it does not describe the event that caused the request.

## Directory map

| Directory | Responsibility |
| --- | --- |
| `chaosimpl/` | `Apply` and `Recover` implementations for top-level chaos kinds. Each implementation is registered as a `ChaosImplPair` through Fx. |
| `common/` | The ordered reconciliation pipeline shared by top-level chaos resources: finalizer initialization, desired phase, conditions, records, and finalizer cleanup. |
| `action/` | Generic action multiplexing used by chaos implementations that dispatch behavior from an action field. |
| `podhttpchaos/`, `podiochaos/`, `podnetworkchaos/` | Reconcilers for the pod-level child CRDs that communicate with Chaos Daemon. |
| `schedule/` | Cron triggering, active-object tracking, garbage collection, and pause propagation for `Schedule`. |
| `statuscheck/` | Reconciliation and in-process workers for periodic status checks. |
| `multicluster/` | Remote-cluster lifecycle, remote chaos creation, and status/finalizer synchronization. |
| `config/` | Controller-manager configuration and `ENABLED_CONTROLLERS` filtering. |
| `types/` | Fx object groups used to register chaos objects and webhook objects. |
| `utils/` | Shared controller builders, recorders, Chaos Daemon clients, and controller helpers. |

Workflow reconcilers live under `pkg/workflow/controllers/` but are registered from `controllers.Module` alongside the controllers in this directory.

## Runtime assembly

`controllers/fx.go` is the top-level composition point. It:

- provides the Chaos Daemon client builder, event recorder builder, common pipeline steps, and remote-cluster registry;
- registers the common chaos pipeline, pod-level controllers, workflow controllers, status checks, and multicluster controllers;
- includes the four schedule controllers; and
- loads every chaos implementation from `chaosimpl.AllImpl`.

The common bootstrap creates one controller named `<chaos-name>-pipeline` for each registered `ChaosImplPair`. It also watches child PodHTTPChaos, PodIOChaos, or PodNetworkChaos resources when the implementation declares them. Top-level objects with a non-empty remote-cluster field are filtered out of the local common pipeline and handled by the multicluster controllers instead.

Controller and webhook enablement comes from `ENABLED_CONTROLLERS` and `ENABLED_WEBHOOKS` in `pkg/config/controller.go`. Keep the controller name passed to `ShouldSpawnController` stable because it is user-facing configuration.

## Design rules

### One writer per field

A field should have one logical controller owner. Multiple reconcilers writing the same field create conflict retries at best and contradictory state transitions at worst.

For the common chaos lifecycle, ownership is intentionally divided as follows:

- `desiredphase` owns `.status.experiment.desiredPhase`;
- `condition` owns `.status.conditions`;
- `records` owns `.status.experiment.containerRecords` and implementation-specific custom status updated with those records;
- `finalizers` owns the common chaos finalizer lifecycle.

Do not update one of these fields from a chaos implementation. `ChaosImpl.Apply` and `ChaosImpl.Recover` should operate on one selected target and return the resulting phase; target selection, iteration, phase transitions, persistence, and retries belong to the common pipeline.

### Make reconciliation idempotent and level-based

Do not infer desired behavior from a specific create, update, or delete event. Re-read the object and its dependencies, compare desired and observed state, and perform only the missing transition.

External side effects must tolerate repeated reconciliation. Before creating, injecting, recovering, or deleting something, check whether the requested state has already been reached whenever the underlying API permits it.

### Keep ordering explicit

Separately registered controllers must not depend on informer delivery order. When ordering is required, express it explicitly, as the common chaos pipeline does. Its current order and result propagation are documented in `common/pipeline/README.md`.

Each controller or pipeline step should still have a small, understandable ownership boundary. If its behavior cannot be summarized clearly, reconsider whether responsibilities or resource boundaries should be split.

### Update only owned state on conflicts

Use Kubernetes conflict retries around writes that can race. Re-fetch the latest object inside the retry and reapply only the field owned by that reconciler. Do not rerun external side effects inside a conflict retry unless they are known to be idempotent.

Use the reconcile `context.Context` for API calls and downstream work so cancellation and deadlines propagate. Avoid introducing new `context.TODO()` calls in reconciliation paths.

## Error and requeue semantics

The repository currently uses controller-runtime v0.21. Its reconcile contract is:

- `return ctrl.Result{}, err`: the result is ignored and a non-terminal error is retried with exponential backoff;
- `return ctrl.Result{RequeueAfter: delay}, nil`: reconcile again after a known delay, suitable for deadlines or explicit polling;
- `return ctrl.Result{}, nil`: reconciliation is complete until a watched event enqueues the object again;
- `return ctrl.Result{}, reconcile.TerminalError(err)`: record the error without retrying; use only when retrying cannot make progress.

Do not return both a non-zero result and a non-nil error because controller-runtime ignores the result.

`ctrl.Result.Requeue` is deprecated in the current controller-runtime version. Existing controllers still contain `Requeue: true` paths, but new code should normally return the retryable error or use an explicit `RequeueAfter`, depending on the intended behavior. Preserve compatibility when changing an existing retry path and add tests for the timing/error semantics.

Treat `NotFound` as successful completion when deletion is the expected explanation. For other API and external-system failures, do not only log and return success unless a future watched event is guaranteed to retry the work.

## Adding or changing controllers

When changing an existing controller:

1. Identify the fields and external side effects it owns.
2. Keep business logic in a testable reconciler or helper rather than the Fx bootstrap.
3. Use `controllers/utils/builder.Default` unless the controller requires custom builder behavior.
4. Gate new top-level controllers with a stable `ShouldSpawnController` name when users need to disable them.
5. Register new providers or bootstraps in the narrowest Fx module, then include that module in `controllers.Module` if necessary.
6. Add focused tests for normal reconciliation, deletion, conflicts, retries, and idempotency as applicable.

For a new top-level chaos kind, also:

1. define and mark the API type under `api/v1alpha1/`;
2. implement single-target `Apply` and `Recover` behavior under `chaosimpl/`;
3. provide a `ChaosImplPair` from the implementation's Fx module and include it in `chaosimpl.AllImpl`;
4. run `make generate` and inspect generated API code, clients, CRDs, workflow/schedule registration, and frontend mappings.

## Focused validation

Use the narrowest package tests during development, for example:

```bash
go test ./controllers/common/...
go test ./controllers/schedule/...
go test ./controllers/statuscheck/...
go test ./controllers/multicluster/...
go test ./controllers/chaosimpl/podchaos/...
```

Some suites use controller-runtime envtest and require its API-server/etcd assets; run them in the repository development environment when those assets are unavailable locally. Reserve `make check` and `make test` for appropriate broad or final validation rather than running them after every small edit.

When API types or generated registration change, run the required generators before interpreting test failures or reviewing the final diff.
