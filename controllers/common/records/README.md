# Records pipeline step

The records step owns the Go field `.Status.Experiment.Records`, serialized as `.status.experiment.containerRecords`. It selects targets, drives each target toward the desired phase, and persists per-target progress.

## Target selection

Selection runs only while the records field is `nil`:

1. Iterate over the named selector specs returned by the chaos object.
2. Resolve targets through the shared selector.
3. Create one record per target with its ID, selector key, and initial phase `Not Injected`.
4. Persist the records with the rest of the reconciliation result.

A nil records field means selection has not completed. The current implementation does not dynamically reselect replacements when selected pods, containers, machines, or volumes disappear.

Each selector must return at least one target. Selection errors or an empty result are recorded as failures, but the current implementation returns without an explicit requeue. Another watched event is required to attempt selection again.

## Phase transitions

For every record, the reconciler compares `record.Phase` with `DesiredPhase` and calls one operation on the registered `ChaosImpl`:

- `Run` drives the record toward `Injected`;
- `Stop` drives the record toward `Not Injected`.

Implementations may return intermediate phases prefixed by `Not Injected` or `Injected`. The records controller preserves the transition cycle and may call the opposite operation first when required to move safely through an intermediate state.

A chaos implementation should operate on one selected record. It must not select targets, iterate over the full target set, or persist the common records field; those responsibilities belong here.

## Persisted state

When state changes, the step persists:

- the record phase;
- successful injection and recovery counters;
- bounded per-record success/failure events, using `MAX_EVENTS`; and
- implementation-specific custom status when the chaos type exposes it.

Writes re-fetch the latest object inside `retry.RetryOnConflict`. The implementation currently returns the legacy `Requeue: true` result when an apply/recover operation needs retry or a records write fails. `Requeue` is deprecated by controller-runtime; preserve behavior carefully when migrating these paths to returned errors or explicit polling.

## Selector and implementation design

Selectors must support targets beyond pods, including containers, volumes, physical machines, and cloud instances. A chaos type can expose multiple named selectors, as NetworkChaos does; `SelectorKey` records which selector produced each target.

Keep `ChaosImpl.Apply` and `ChaosImpl.Recover` small, idempotent where possible, and focused on the target identified by the supplied record index.

Focused validation:

```bash
go test ./controllers/common/...
go test ./controllers/chaosimpl/podchaos/...
```
