---
id: networkchaos_experiment
title: NetworkChaos Experiment
sidebar_label: NetworkChaos Experiment
---

This document describes how to create NetworkChaos experiments in Chaos Mesh.

NetworkChaos actions are divided into two categories:

- **Network Partition** action separates pods into several independent subnets by blocking communication between them.

- **Network Emulation (Netem) Chaos** actions cover regular network faults, such as network delay, duplication, loss, and corruption.

## Network Partition Action

Below is a sample network partition configuration file:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: network-partition-example
  namespace: chaos-testing
spec:
  action: partition
  mode: one
  selector:
    labelSelectors:
      "app.kubernetes.io/component": "tikv"
  direction: to
  target:
    selector:
      namespaces:
        - tidb-cluster-demo
      labelSelectors:
        "app.kubernetes.io/component": "tikv"
    mode: one
  duration: "10s"
  scheduler:
    cron: "@every 15s"
```

For more sample files, see [examples](https://github.com/chaos-mesh/chaos-mesh/tree/master/examples). You can edit them as needed.

Description:

* **action** defines the specific chaos action for the pod. In this case, it is network partition.
* **mode** defines the mode to run chaos action.
* **selector** specifies the target pods for chaos injection. For more details, see [Define the Scope of Chaos Experiment](../user_guides/experiment_scope.md).
* **direction** specifies the partition direction. Supported directions are `from`, `to` and `both`.
* **target** specifies the target for network partition.
* **duration** defines the duration for each chaos experiment. In the sample file above, the network partition lasts for `10` seconds.
* **scheduler** defines the scheduler rules for the running time of the chaos experiment. For more rule information, see [robfig/cron](https://godoc.org/github.com/robfig/cron).

## Netem Chaos Actions

There are 4 cases for netem chaos actions, namely loss, delay, duplicate, and corrupt.

> **Note:**
>
> The detailed description of each field in the configuration template are consistent with that in [Network Partition](#network-partition-action).

### Network Loss

A Network Loss action causes network packets to drop randomly. To add a Network Loss action, locate and edit the corresponding template in [/examples](https://github.com/chaos-mesh/chaos-mesh/blob/master/examples/network-loss-example.yaml).

In this case, two action specific attributes are required - loss and correlation.

```yaml
loss:
  loss: "25"
  correlation: "25"
```

**loss** defines the percentage of packet loss.

NetworkChaos variation is not purely random, so to emulate that there is a correlation value as well.

### Network Delay

A Network Delay action causes delays in message sending. To add a Network Delay action, locate and edit the corresponding template in [/examples](https://github.com/chaos-mesh/chaos-mesh/blob/master/examples/network-delay-example.yaml).

In this case, three action specific attributes are required - correlation, jitter, and latency.

```yaml
delay:
  latency: "90ms"
  correlation: "25"
  jitter: "90ms"
```

**latency** defines the delay time in sending packets.

**jitter** specifies the jitter of the delay time. Default is `0ms`.

**correlation** specifies the correlation of the jitter. Default is `0`.

In the above example, the network latency is 90ms ± 90ms with 25% correlation.

### Network Duplicate

A Network Duplicate action causes packet duplication. To add a Network Duplicate action, locate and edit the corresponding template in [/examples](https://github.com/chaos-mesh/chaos-mesh/blob/master/examples/network-duplicate-example.yaml).

In this case, two attributes are required - correlation and duplicate.

```yaml
duplicate:
  duplicate: "40"
  correlation: "25"
```

**duplicate** indicates the percentage of packet duplication. In the above example, the duplication rate is 40%.

### Network Corrupt

A Network Corrupt action causes packet corruption. To add a Network Corrupt action, locate and edit the corresponding template in [/examples](https://github.com/chaos-mesh/chaos-mesh/blob/master/examples/network-corrupt-example.yaml).

In this case, two action specific attributes are required - correlation and corrupt.

```yaml
corrupt:
  corrupt: "40"
  correlation: "25"
```

**corrupt** specifies the percentage of packet corruption.

## Network Bandwidth Action

Network Bandwidth Action is used to limit the network bandwidth. To add a Network Bandwidth Action, locate and edit the corresponding template in [/examples](https://github.com/chaos-mesh/chaos-mesh/blob/master/examples/network-bandwidth-example.yaml).

> **Note:**
>
> Minikube does not support this feature as `CONFIG_NET_SCH_TBF` is disabled in Minikube's image.

To inject Network Bandwidth fault, three action specific attributes are required - rate, buffer and limit.

```yaml
 bandwidth:
   rate: 10 kbps
   buffer: 1000
   limit: 100
```

**rate** allows "bps", "kbps", "mbps", "gbps", "tbps" unit. "bps" means bytes per second.

**limit** defines the number of bytes that can be queued waiting for tokens to become available.

**buffer** is the maximum amount of bytes that tokens can be available for instantaneously.

**peakrate** is the maximum depletion rate of the bucket.

**minburst** specifies the size of the peakrate bucket.
