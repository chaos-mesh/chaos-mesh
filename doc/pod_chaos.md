# Pod Chaos Document

This document helps you to build pod chaos experiments. 

> ** note ** 
> 
> Chaos mash does not currently support simulation injection of naked pods. And it only supports some specific pods, such as `deployment` 、`statefulset` 、`daemonset`. 

Pod chaos allows you to simulate pod faults, specifically `pod failure` and `pod kill`.

- **Pod Failure** action periodically injects errors to pods. And it will cause the pod to not be created for a while.

- **Pod Kill** action kills the specified pod.

## Pod Failure Action

Below is a sample pod failure configuration file:

```yaml
apiVersion: pingcap.com/v1alpha1
kind: PodChaos
metadata:
  name: pod-failure-example
  namespace: chaos-testing
spec:
  action: pod-failure
  mode: one
  duration: "30s"
  selector:
    namespaces:
      - tidb-cluster-demo
    labelSelectors:
      "app.kubernetes.io/component": "tikv"
  scheduler:
    cron: "@every 2m"
```

> For more sample files, see [examples](../examples). You can edit them as needed. 

Description:

* **action** defines the specific chaos action for the pod. In this case, it is pod failure.
* **mode** defines the mode to run chaos action. Supported mode: `one / all / fixed / fixed-percent / random-max-percent`.
* **duration** defines the duration for each chaos experiment. In the sample file above, the pod failure lasts for 30 seconds.
* **selector** is used to select pods that are used to inject chaos actions.
* **scheduler** defines the scheduler rules for the running time of the chaos experiment. For more rule information, see <https://godoc.org/github.com/robfig/cron>.

## Pod Kill Action

> **Note:** 
> 
> The detailed description of each field in the configuration template are consistent with that in [Pod Failure](#Pod-Failure-Action).

Below is a sample pod kill configuration file:

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

> For more sample files, see [examples](../examples). You can edit them as needed. 
