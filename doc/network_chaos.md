# Network Chaos Document

This document describes how to add network chaos experiments in Chaos Mesh.

Network chaos are mainly divided into two categories, namely **netem chaos** and **network partition**.

Netem chaos contains some kinds of network chaos, such as delay, duplication, loss and corrupt.

Network partition can decompose pods into several independent subnets by blocking communication between them.

## Network Partition Action
Sample network partition ducument:
```yaml
apiVersion: pingcap.com/v1alpha1
kind: NetworkChaos
metadata:
  name: network-partition-example
  namespace: chaos-testing
spec:
  action: partition
  mode: one
  selector:
    namespaces:
      - tidb-cluster-demo
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
User can find and edit the template refer to [examples/network-partition-example.yaml](../examples/network-partition-example.yaml).
* **action** defines the specific pod chaos action. In this case, it means network partition, represents the chaos action of network partition of pods.
* **mode** defines the mode to run chaos action.
* **selector** is used to select pods that are used to inject chaos action.
* **direction** represents the partition direction, supported direction: from / to / both.
* **target** represents network partition target.
* **duration** define the duration time for each chaos experiment. As the example shows, the network partition lasts 10 seconds.
* **scheduler** defines some scheduler rules to the running time of the chaos experiment about pods. More rule info: https://godoc.org/github.com/robfig/cron


## Netem Chaos Actions

There are 4 cases, loss, delay, duplicate and corrupt.

The meanings of action, mode, selector duration, scheduler are consistent with the description in the Network Partition.

### Network Loss

Network Loss means that network packets are dropped randomly.
> In this case, two attributes are required, loss and correlation.
>
> ```yaml
> loss:
>   loss: "25"
>   correlation: "25"
> ```
> Loss represents the percentage of packet loss. The above example shows a 25% chance of packet loss.
>
> Network chaos variation isn't purely random, so to emulate that there is a correlation value as well.

### Network Delay

Network Delay means to delay the sending of network messages.
> In this case, three attributes are required, correlation, jitter and latency.
>
>```yaml
>  delay:
>    latency: "90ms"
>    correlation: "25"
>    jitter: "90ms"
>```
> Latency indicates the delay time in sending packets.
>
> jitter represents the jitter of the delay time.
>
> The above example shows that the network latency is 90ms ± 90ms.

### Network Duplicate

Network duplicate means packet duplication.
> In this case, two attributes are required, correlation and duplicate.
>
>```yaml
>  duplicate:
>    duplicate: "40"
>    correlation: "25"
>```
>
> Network duplicate is specified the same way as network loss. The parameter "Duplicate" indicates the percentage of packet duplication. And it shows that duplication rate is 40%. 

### Network Corrupt

Network corrupt means packet corruption.
> In this case, two attributes are required, correlation and corrupt.
>
>```yaml
>  corrupt:
>    corrupt: "40"
>    correlation: "25"
>```
>
> Similar to the other cases described above, the parameter "corrupt" indicates the percentage of packet corruption.
