# Stress Chaos Document

This document helps you to build stress chaos experiments.

Stress chaos is a chaos to generate plenty of stresses over a collection of pods. A sidecar will be injected along with the target pod during creating. It's the sidecar which generates stresses or cancels them. For now, we use `stress-ng` as the stress generator for the chaos.

## Usage 

### Define a `StressChaos`

An example `yaml` of `StressChaos` which burns 1 CPU for 30 seconds in every 2 minutes is listed below. You can specify all of `stress-ng`  stressors via the `stressors` field, see `stress-ng(1)` for supported stressors. The others `spec` fields are common to all of chaos. 

```yaml
apiVersion: pingcap.com/v1alpha1
kind: StressChaos
metadata:
  name: burn-cpu
  namespace: chaos-testing
spec:
  mode: one
  selector:
    namespaces:
      - tidb-cluster-demo
  stressors: "--cpu 1"
  duration: "30s"
  scheduler:
    cron: "@every 2m"
```

### Inject a `StressChaos`

```bash
# If you do not have a namespace for your application, you can create one such as tidb-cluster-demo
kubectl create ns tidb-cluster-demo
# Inject a stress chaos
kubectl apply -f your-stress-chaos.yaml
# Your pod's cpu will burn for 30s
```

Have fun : )

