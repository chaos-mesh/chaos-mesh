---
id: podchaos_experiment
title: PodChaos Experiment
sidebar_label: PodChaos Experiment
---

This document introduces how to create PodChaos experiments.

> **Note:**
>
> Currently, Chaos Mesh does not support simulation injection of naked pods. And it only supports some specific pods, such as `deployment`, `statefulset`, `daemonset`.

PodChaos allows you to simulate pod faults, specifically `pod failure` and `pod kill`. `pod failure` can be used to simulate a situation where a pod is down. In this case, the pod is unavailable for a long time.

- **Pod Failure** action periodically injects errors to pods. And it will cause pod creation failure for a while. In other words, the selected pod will be unavailable in a specified period.

- **Pod Kill** action kills the specified pod (ReplicaSet or something similar might be needed to ensure the pod will be restarted).

## `pod-failure` configuration file

Below is a sample `pod-failure` configuration file:

```yaml
apiVersion: pingcap.com/v1alpha1
kind: PodChaos
metadata:
  name: pod-failure-example
  namespace: chaos-testing
spec:
  action: pod-failure
  mode: one
  value: ""
  duration: "30s"
  selector:
    labelSelectors:
      "app.kubernetes.io/component": "tikv"
  scheduler:
    cron: "@every 2m"
```

For more sample files, see [examples](https://github.com/pingcap/chaos-mesh/tree/master/examples). You can edit them as needed.

Description:

* **action** defines the specific chaos action for the pod. In this case, it is pod failure.
* **mode** defines the mode to run chaos action. Supported mode: `one` / `all` / `fixed` / `fixed-percent` / `random-max-percent`.
* **value** depends on the value of `mode`. If `mode` is `one` or `all`, leave `value` empty. If `fixed`, provide an integer of pods to do chaos action. If `fixed-percent`, provide a number from 0 to 100 to specify the percent of pods the server can do chaos action. If `random-max-percent`, provide a number from 0 to 100 to specify the max percent of pods to do chaos action.
* **duration** defines the duration for each chaos experiment. The value of the `duration` field is `30s`, which indicates that pod failure will last 30 seconds.
* **selector** is used to select pods that are used to inject chaos actions.
* **scheduler** defines the scheduler rules for the running time of the chaos experiment. For more rule information, see <https://godoc.org/github.com/robfig/cron>.

## `pod-kill` configuration file

Below is a sample `pod-kill` configuration file:

```yaml
apiVersion: pingcap.com/v1alpha1
kind: PodChaos
metadata:
  name: pod-kill-example
  namespace: chaos-testing
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - tidb-cluster-demo
    labelSelectors:
      "app.kubernetes.io/component": "tikv"
  scheduler:
    cron: "@every 1m"
```

The detailed description of each field in the configuration template are consistent with that in [`pod-failure`](#pod-failure-configuration-file).
