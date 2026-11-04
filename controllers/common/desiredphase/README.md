# Desired phase pipeline step

The desired-phase step is the sole owner of the Go field `.Status.Experiment.DesiredPhase`, serialized as `.status.experiment.desiredPhase`.

It does not inject or recover targets. It decides whether the records step should drive every target toward `Injected` or `Not Injected`.

## Decision order

The reconciler calculates the field on every run in this order:

1. A deleting object is `Stop`.
2. A one-shot experiment is `Run`. Except for deletion, one-shot experiments ignore duration and pause so the operation is not applied multiple times.
3. An experiment whose duration has expired is `Stop`.
4. An object with `experiment.chaos-mesh.org/pause: "true"` is `Stop`.
5. Every other object is `Run`.

The API values are `Run` and `Stop`, not “running” and “stopped”.

When a duration is configured and has not expired, the step returns `RequeueAfter` for the remaining duration. The pipeline retains that deadline while continuing through later steps.

## Persistence and events

If the calculated value differs from the persisted value, the reconciler:

- emits the corresponding started, paused, time-up, or deleted event;
- re-fetches the object inside `retry.RetryOnConflict`;
- updates only `DesiredPhase`; and
- emits an updated or failed event.

The records step later observes the desired phase and performs the actual `Apply` or `Recover` transitions.

Focused validation:

```bash
go test ./controllers/common/...
```
