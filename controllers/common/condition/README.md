# Conditions pipeline step

The conditions step is the sole owner of `.status.conditions` for common chaos resources. It derives user-facing summary conditions from the pause annotation and the records currently persisted on the object; it does not select targets or perform injection/recovery.

## Derived conditions

Each reconciliation replaces the common condition set with these four condition types:

| Type | True when |
| --- | --- |
| `Selected` | `.status.experiment.containerRecords` is non-nil. |
| `AllInjected` | Records are non-nil and every record has phase `Injected`. |
| `AllRecovered` | Records are non-nil and every record has phase `Not Injected`. |
| `Paused` | `experiment.chaos-mesh.org/pause` is `"true"`. |

A non-nil empty record slice satisfies both “all” checks. In normal flow, target selection does not persist an empty slice, but tests and new selector behavior should preserve the intended distinction between nil and empty records.

The current implementation does not populate a reason for these common conditions.

## Pipeline timing

This step runs before the records step. Conditions therefore reflect the records read from the API server at the beginning of the pipeline run. When records later change, that status update triggers another reconciliation and the conditions converge on the new phases.

Do not make this step inject, recover, or change desired phase. If a new condition depends on additional state, keep the derivation level-based and update the condition tests.

## Persistence

The reconciler compares conditions by type, status, and reason. When they differ, it re-fetches the object inside `retry.RetryOnConflict`, replaces `.status.conditions`, and updates the object.

Focused validation:

```bash
go test ./controllers/common/condition/...
go test ./controllers/common/...
```
