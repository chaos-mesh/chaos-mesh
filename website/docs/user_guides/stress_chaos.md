---
id: stresschaos_experiment
title: StressChaos Experiment
sidebar_label: StressChaos Experiment
---

This document helps you create StressChaos experiments.

StressChaos can generate plenty of stresses over a collection of pods. The stressors is injected into the target pods via the `chaos-daemon` internally.

## Configuration

A StressChaos shares common configurations like other chaos, such as how to select pods, how to specify periodic chaos. You can refer to other docs for details. It defines stressors in **either** of the following two ways:

* `stressors`

  `Stressors` defines a plenty of stressors supported to stress system components out. You can use one or more of them to make up various kinds of stresses. At least one of the stressors should be specified. The following is supported stressors for now:

  1. `memory`

     A `memory` stressor will continuously stress virtual memory out.

     | Option    | Type    | Required | Description                                                  |
     | --------- | ------- | -------- | ------------------------------------------------------------ |
     | `workers` | Integer | True     | Specifies concurrent stressing instance.                      |
     | `size`   | String  | False    | Specifies memory size consumed per worker, default is the total available memory. One can also specify the size as *%* of total available memory or in units of *B, KB/KiB, MB/MiB, GB/GiB, TB/TiB*. |

  2. `cpu`

     A `cpu` stressor will continuously stress CPU out.

     | Option    | Type    | Required | Description                                                  |
     | --------- | ------- | -------- | ------------------------------------------------------------ |
     | `workers` | Integer | True     | Specifies concurrent stressing instance. Actually it specifies how many CPUs to stress when it's less than available CPUs. |
     | `load`    | Integer | False    | Specifies  percent loading per worker. 0 is effectively a sleep (no load) and 100 is full loading. |

* `stressngStressors`

    `StressngStressors` defines a plenty of stressors just like `Stressors` except that it's an experimental feature and more powerful.

    You can define stressors in `stress-ng` (see also `man stress-ng`) dialect.

    > **Note:**
    >
    > However, not all of the supported stressors are well tested. It might be retired in later releases. Therefore, it is recommended to use `Stressors` to define the stressors and use this only when you want more stressors unsupported by `Stressors`.

    When both `StressngStressors` and `Stressors` are defined, `StressngStressors` wins.

## Usage

Below is an example YAML file of StressChaos which is set to burn 1 CPU for 30 seconds in every 2 minutes:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: burn-cpu
  namespace: chaos-testing
spec:
  mode: one
  selector:
    namespaces:
      - tidb-cluster-demo
  stressors:
    cpu:
      workers: 1
  duration: "30s"
  scheduler:
    cron: "@every 2m"
```

1. Create a namespace for your application. For example, `tidb-cluster-demo`:

    ```bash
    kubectl create ns tidb-cluster-demo
    ```

2. Create your pods in the target namespace:

    ```bash
    kubectl apply -f *your-pods.yaml*
    ```

3. Inject a StressChaos:

    ```bash
    kubectl apply -f *your-stress-chaos.yaml*
    ```

Then, your pod's CPU will burn for 30 seconds.
