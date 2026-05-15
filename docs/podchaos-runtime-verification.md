# PodChaos runtime verification walkthrough

Purpose
-------
This short guide shows practical steps to verify the runtime effects of a `PodChaos` experiment. It's aimed at beginners who want reliable ways to observe and fact-check the impact of Pod-level chaos (restarts, kills, lifecycle faults).

Prerequisites
-------------
- `chaos-mesh` is installed and the CRDs are available.
- `kubectl` configured to talk to the target cluster and namespace.
- A workload (pods) labeled so it can be targeted by `PodChaos`.

Run a PodChaos experiment
-------------------------
Use an existing sample or a small manifest. Example using the repository sample:

```sh
kubectl apply -f examples/pod-kill-example.yaml
```

Or apply a minimal `PodChaos` manifest that targets your app namespace/labels.

Observe runtime impact
----------------------
- Watch pods in real time:

```sh
kubectl get pods -w -n <target-namespace>
```

- Inspect container restart counts and termination reasons:

```sh
kubectl describe pod <pod-name> -n <target-namespace>
```

- Check the `PodChaos` resource status and spec to confirm what was requested:

```sh
kubectl describe podchaos <podchaos-name> -n <target-namespace>
kubectl get podchaos <podchaos-name> -n <target-namespace> -o yaml
```

Correlating restarts/events with runtime impact
-----------------------------------------------
- Note timestamps when the `PodChaos` action is applied (watch `kubectl apply` output or `kubectl get events`).
- Match pod `Last State` / `State` timestamps and `restartCount` in `kubectl describe pod` output to the chaos action time.
- Inspect application logs for the same time window:

```sh
kubectl logs <pod-name> -c <container> -n <target-namespace> --since=5m
```

- Check cluster events for eviction, OOM, or Kubelet actions that explain the restarts:

```sh
kubectl get events -n <target-namespace> --sort-by='.lastTimestamp'
```

Limitations of relying only on PodChaos status
---------------------------------------------
- `PodChaos` resource status may be minimal or delayed and does not prove end-to-end application-level impact.
- Pod restart counts and Kubernetes events show low-level outcomes but do not show functional correctness or request-level failures.
- For strong verification, combine PodChaos observation with:
  - application logs and error rates
  - service-level metrics (latency, success rates)
  - distributed traces where available

Quick verification checklist
---------------------------
- Apply the `PodChaos` manifest and note the timestamp.
- `kubectl get pods -w` to see immediate pod restarts/terminations.
- `kubectl describe podchaos` and `kubectl get podchaos -o yaml` to confirm the targeted selector and action.
- `kubectl describe pod` + `kubectl logs` to correlate restart counts and application behavior.
- Review events and metrics to support the observed effects.

Notes for contributors
----------------------
Keep experiments reproducible and small: prefer targeting a single pod or a narrow selector for verification. Document the exact manifest used and the namespace so others can reproduce the same fact-check.

Related examples
----------------
See `examples/pod-kill-example.yaml` and `examples/pod-failure-example.yaml` in this repository for ready-to-run PodChaos examples.
