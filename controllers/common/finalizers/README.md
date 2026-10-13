# Finalizer pipeline steps

The finalizer package owns the common chaos finalizer lifecycle. It contributes two steps to the common pipeline so initialization happens before any injection and cleanup happens after record recovery.

The common finalizer is:

```text
chaos-mesh/records
```

## Initialization

`InitReconciler` runs first. For a live object, it adds `chaos-mesh/records` when that value is absent. It does not add a finalizer after deletion has started.

This prevents Kubernetes from deleting the chaos object before the records step has an opportunity to recover injected targets.

## Cleanup

`CleanReconciler` runs last and acts only on a deleting object. Cleanup is allowed when either:

- every record has phase `Not Injected`; or
- the object has the following forced-cleanup annotation:

```yaml
chaos-mesh.chaos-mesh.org/cleanFinalizer: forced
```

A nil or empty record list is considered fully recovered because there is no recorded target to recover.

Forced cleanup bypasses the record-phase guard in `CleanReconciler` and can leave faults behind. It does not bypass earlier pipeline steps. If record recovery or a records write keeps returning a requeue, the pipeline stops before cleanup and the annotation cannot remove the finalizers.

## Current finalizer-list behavior

The current cleanup implementation calls `SetFinalizers([]string{})`; it clears the entire finalizer list rather than removing only `chaos-mesh/records`. Do not describe this behavior as preserving unrelated finalizers.

Changing cleanup to preserve third-party finalizers is compatibility-sensitive. Add tests covering multiple finalizers, normal deletion, partial recovery, empty records, and forced cleanup before changing it.

Both initialization and cleanup re-fetch the latest object inside `retry.RetryOnConflict` before writing finalizers and emit lifecycle events through the shared recorder.

Focused validation:

```bash
go test ./controllers/common/...
```
