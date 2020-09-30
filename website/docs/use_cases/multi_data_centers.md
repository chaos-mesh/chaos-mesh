---
id: multi_data_centers
title: Network latency simulation across multiple data centers
sidebar_label: Network latency simulation across multiple data centers
---

This document helps you simulate multiple data centers scenarios.

## Characteristics of multiple data centers scenarios

- The latency between different data centers
- The bandwidth limitations between data centers

> **Note**:
>
> Currently, Chaos Mesh cannot simulate the scenario of the bandwidth limitations between data centers. So in this case, only simulate the scenario of the latency between different data centers.

## Experiment environment

Suppose our application will be deployed in three data centers in a production environment
and these data centers are still under construction. Now we want to test the impact of
such a deployment topology on the business in advance.

Here we use TiDB cluster as an example. Suppose we already install the [TiBD cluster](https://docs.pingcap.com/tidb-in-kubernetes/stable/) and [Chaos Mesh](get_started/installation.md)
in our Kubernetes environment. In this TiDB cluster, we have three TiDB pods, three PD pods and seven TiKV pods:

```bash
kubectl get pod -n tidb-cluster # "tidb-cluster" is the namespace of TiDB cluster
```

Output:

```bash
NAME                               READY   STATUS    RESTARTS   AGE
basic-discovery-7f9f48c465-6pdhn   1/1     Running   0          30m
basic-pd-0                         1/1     Running   0          30m
basic-pd-1                         1/1     Running   0          30m
basic-pd-2                         1/1     Running   0          30m
basic-tidb-0                       2/2     Running   0          29m
basic-tidb-1                       2/2     Running   0          29m
basic-tidb-2                       2/2     Running   0          29m
basic-tikv-0                       1/1     Running   0          29m
basic-tikv-1                       1/1     Running   0          29m
basic-tikv-2                       1/1     Running   0          29m
basic-tikv-3                       1/1     Running   0          29m
basic-tikv-4                       1/1     Running   0          29m
basic-tikv-5                       1/1     Running   0          29m
basic-tikv-6                       1/1     Running   0          29m
```

### Grouping

`dc-a`, `dc-b`, and `dc-c` are the three data centers we will use later. So we will split the pods to these data centers:

|      dc-a      |      dc-b      |       dc-c       |
| :------------: | :------------: | :--------------: |
|   basic-pd-0   |   basic-pd-1   |    basic-pd-2    |
|  basic-tidb-0  |  basic-tidb-1  |   basic-tidb-2   |
| basic-tikv-0/1 | basic-tikv-2/3 | basic-tikv-4/5/6 |

### Latency between three data centers

|                | latency |
| :------------: | :-----: |
| dc-a <--> dc-b |   1ms   |
| db-a <--> dc-c |   2ms   |
| dc-b <--> dc-c |   2ms   |

## Inject network latency

### Design injection rules

Chaos Mesh provides [`NetworkChaos`](chaos_experiments/network_chaos.md) to inject network latency,
so we can use it to simulate the latency between three data centers.

At present, `NetworkChaos` has a limitation that each target pod only has one configuration of `netem` in effect.
So we can use the following rules:

| source pods | latency | target pods |
| :---------: | :-----: | :---------: |
|    dc-a     |   1ms   |    dc-b     |
|    dc-a     |   1ms   |    dc-c     |
|    dc-b     |   1ms   |    dc-c     |
|    dc-c     |   1ms   |    dc-a     |
|    dc-c     |   1ms   |    dc-b     |

According to above rules, the latency between `dc-a` and `dc-b` is `1ms`, the latency between `dc-a` and `dc-c` is `2ms`
and the latency between `dc-b` and `dc-c` is `2ms`.

### Define the chaos experiment

According to the injection rules, we define the chaos experiment as following:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: network-delay-a
  namespace: tidb-cluster
spec:
  action: delay # chaos action
  mode: all
  selector: # define the pods belong to dc-a
    pods:
      tidb-cluster: # namespace of the target pods
        - basic-tidb-0
        - basic-pd-0
        - basic-tikv-0
        - basic-tikv-1
  delay:
    latency: "1ms"
  direction: to
  target:
    selector: # define the pods belong to dc-b and dc-c
      pods:
        tidb-cluster: # namespace of the target pods
          - basic-tidb-1
          - basic-tidb-2
          - basic-pd-1
          - basic-pd-2
          - basic-tikv-2
          - basic-tikv-3
          - basic-tikv-4
          - basic-tikv-5
          - basic-tikv-6
    mode: all

---
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: network-delay-b
  namespace: tidb-cluster
spec:
  action: delay
  mode: all
  selector: # define the pods belong to dc-b
    pods:
      tidb-cluster: # namespace of the target pods
        - basic-tidb-1
        - basic-pd-1
        - basic-tikv-2
        - basic-tikv-3
  delay:
    latency: "1ms"
  direction: to
  target:
    selector: # define the pods belong to dc-c
      pods:
        tidb-cluster: # namespace of the target pods
          - basic-tidb-2
          - basic-pd-2
          - basic-tikv-4
          - basic-tikv-5
          - basic-tikv-6
    mode: all

---
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: network-delay-c
  namespace: tidb-cluster
spec:
  action: delay
  mode: all
  selector: # define the pods belong to dc-c
    pods:
      tidb-cluster: # namespace of the target pods
        - basic-tidb-2
        - basic-pd-2
        - basic-tikv-4
        - basic-tikv-5
        - basic-tikv-6
  delay:
    latency: "1ms"
  direction: to
  target:
    selector: # define the pods belong to dc-a and dc-b
      pods:
        tidb-cluster: # namespace of the target pods
          - basic-tidb-0
          - basic-tidb-1
          - basic-pd-0
          - basic-pd-1
          - basic-tikv-0
          - basic-tikv-1
          - basic-tikv-2
          - basic-tikv-3
    mode: all
```

### Apply the chaos experiment

Define the above chaos experiment as `delay.yaml` and apply this file:

```bash
kubectl apply -f delay.yaml
```

### Check the result

Use `ping` command to check the latency between three centers.

#### Check the latency between the pods belong to `dc-a`

```bash
kubectl exec -it -n tidb-cluster basic-tidb-0 -c tidb -- ping -c 2 basic-tikv-0.basic-tikv-peer.tidb-cluster.svc
```

output:

```bash
PING basic-tikv-0.basic-tikv-peer.tidb-cluster.svc (10.244.1.229): 56 data bytes
64 bytes from 10.244.1.229: seq=0 ttl=63 time=0.095 ms
64 bytes from 10.244.1.229: seq=1 ttl=63 time=0.100 ms
```

From the output, we can see that the latency between the pods belong to `dc-a` is around `0.1ms`.

#### Check the latency between `dc-a` and `dc-c`

```bash
kubectl exec -it -n tidb-cluster basic-tidb-0 -c tidb -- ping -c 2 basic-tidb-1.basic-tidb-peer.tidb-cluster.svc
```

output:

```bash
PING basic-tidb-1.basic-tidb-peer.tidb-cluster.svc (10.244.3.3): 56 data bytes
64 bytes from 10.244.3.3: seq=0 ttl=62 time=1.193 ms
64 bytes from 10.244.3.3: seq=1 ttl=62 time=1.201 ms
```

From the output, we can see that the latency between `dc-a` and `dc-c` is around `1ms`.

#### Check the latency between `dc-b` and `dc-c`

```bash
kubectl exec -it -n tidb-cluster basic-tidb-0 -c tidb -- ping -c 2 basic-tidb-2.basic-tidb-peer.tidb-cluster.svc
```

output:

```bash
PING basic-tidb-2.basic-tidb-peer.tidb-cluster.svc (10.244.2.27): 56 data bytes
64 bytes from 10.244.2.27: seq=0 ttl=62 time=2.200 ms
64 bytes from 10.244.2.27: seq=1 ttl=62 time=2.251 ms
```

From the output, we can see that the latency between `dc-a` and `dc-c` is around `2ms`.

## Delete the network latency

```bash
kubectl delete -f delay.yaml
```
