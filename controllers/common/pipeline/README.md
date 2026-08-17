# Common chaos reconciliation pipeline

Chaos Mesh originally registered the common chaos reconcilers independently. Informer delivery could run them in different orders and produce invalid transitions, including the race described in [issue #2449](https://github.com/chaos-mesh/chaos-mesh/issues/2449). The pipeline makes the dependency order explicit while exposing one controller-runtime reconciler for each top-level chaos kind.

## Assembly

`common.Bootstrap` creates a `PipelineContext` for every registered `ChaosImplPair`, adds the steps returned by `common.AllSteps`, and registers the resulting controller as `<chaos-name>-pipeline`.

The context shared by the steps contains:

- the chaos object type and its `ChaosImpl`;
- the controller-runtime manager, cached client, and no-cache reader;
- the selector and recorder builders; and
- the logger.

The current step order is an invariant:

| Order | Step | Owned state or effect |
| --- | --- | --- |
| 1 | `finalizers.InitStep` | Adds the common records finalizer to a live chaos object. |
| 2 | `desiredphase.Step` | Derives `Run` or `Stop` from deletion, one-shot behavior, duration, and pause state. |
| 3 | `condition.Step` | Derives the selected, injected, recovered, and paused conditions from the currently persisted object. |
| 4 | `records.Step` | Selects targets and drives each record toward the desired phase through `Apply` or `Recover`. |
| 5 | `finalizers.CleanStep` | Removes finalizers after deletion recovery is complete or forced cleanup is requested. |

Do not reorder these steps without analyzing creation, pause, duration, partial injection, recovery, and deletion. In particular, cleanup must run after records so a deleting object is not released before recovery has completed.

The condition step runs before the records step and therefore describes the records persisted at the start of that pipeline run. A records update enqueues another reconciliation, allowing conditions to converge to the new record phases.

## Result propagation

`Pipeline.Reconcile` runs steps sequentially and combines their controller-runtime results:

- a non-nil error stops the pipeline immediately and is returned;
- `Requeue: true` stops the pipeline immediately and requests a rate-limited retry;
- `RequeueAfter` does not stop later steps; the pipeline remembers the earliest requested deadline;
- if an accumulated deadline expires while a later step is running, the pipeline returns an immediate retry without running the remaining steps; and
- after all steps finish, the pipeline returns the earliest remaining `RequeueAfter`.

`Requeue` remains in this implementation for compatibility but is deprecated by the current controller-runtime. New step behavior should prefer returning a retryable error or an explicit `RequeueAfter` as described in the root controller guide.

## Enablement caveat

Several step factories call `ShouldSpawnController`. `Pipeline.AddSteps` stops adding steps when a factory returns `nil`; it does not skip only that step. Selectively disabling a middle common stage can therefore omit every later stage.

Treat the common stages as one ordered unit unless intentionally testing a partial pipeline. If enablement behavior changes, update `controllers/common/step.go`, `Pipeline.AddSteps`, and the common pipeline tests together.

## Watches and predicates

The common controller watches the top-level chaos object and, where declared by its `ChaosImplPair`, the PodHTTPChaos, PodIOChaos, or PodNetworkChaos child objects referenced by records.

Its predicates:

- ignore updates that only append record events, avoiding event-only reconcile loops;
- allow relevant child-CRD updates to enqueue their owning top-level chaos; and
- exclude top-level objects targeting a remote cluster, which are handled by the multicluster controllers.

## Development

Keep step constructors in `step.go` and reconciliation behavior in the owning package. Add or remove steps only through `common.AllSteps`; do not register a second independent controller for a field already owned by the pipeline.

Focused validation:

```bash
go test ./controllers/common/...
```
