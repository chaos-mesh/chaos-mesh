# Stress Chaos Document

This document helps you to build stress chaos experiments.

Stress chaos is a chaos to generate plenty of stresses over a collection of pods. A sidecar will be injected along with the target pod during creating. It's the sidecar which generates stresses or cancels them. For now, we use `stress-ng` as the stress generator for the chaos.

> Note: Stress chaos can only be used if the relevant labels and annotations are set before the application is created. More info refer [here](#create-a-chaos-experiment)

## Usage 

### Configure a ConfigMap

Chaos Mesh uses a sidecar container which is defined in a `ConfigMap` to inject stress chaos.  Before you start the chaos, make sure that the `ConfigMap` is already properly created. An example `ConfigMap` is listed below.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: chaos-stress
  # Same namespace with Chaos Mesh
  namespace: chaos-testing
  labels:
    # Marked to load by Chaos Mesh automatically
    app.kubernetes.io/component: webhook
data:
  chaos-stress.yaml: |
    name: chaos-stress
    containers:
    - name: stress-server
      image: pingcap/chaos-stress:latest
      imagePullpolicy: Always
      ports:
      - containerPort: 65533
      securityContext:
        privileged: true
      command:
        - /usr/local/bin/chaos-stress
        - -addr=:65533
```

The sidecar container contains a `stress-server` which delegates the requested stressors to `stress-ng`. It communicates with the controller via `grpc` specified via `-addr` argument. The container should also export that address (`containerPort` in `chaos-stress.yaml`. When you defined your `ConfigMap` for your application, you can apply it via `kubectl apply -f your-config-map.yaml`

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

Chaos Mesh uses admission-webhook to inject the `stress-server` sidecar container. You should enable it and annotate the application as following.

```bash
# If you do not have a namespace for your application, you can create one such as tidb-cluster-demo
kubectl create ns tidb-cluster-demo
# Enabel admission-webhook if not
kubectl label ns tidb-cluster-demo admission-webhook=enabled
# Set annotation for looking for the sidecar ConfigMap
kubectl annotate ns app-ns admission-webhook.pingcap.com/init-request=chaos-stress
# Create your application
kubectl apply -f your-app.yaml
# Inject a stress chaos
kubectl apply -f your-stress-chaos.yaml
```

Have fun : )

