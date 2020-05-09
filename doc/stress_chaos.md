# Stress Chaos Document

This document helps you to build stress chaos experiments.

Stress chaos is a chaos to generate plenty of stresses over a collection of pods. The stressors are injected into the target pods via the `chaos-daemon` internally. 

A `StressChaos` shares common configurations like other chaos, such as how to select pods, how to specify periodic chaos ... (You can refer to other docs for how to use them). It defines stressors in **either** of the following two ways:

* `stressors`

  Stressors define plenty of stressors supported to stress system components out. You can use one or more of them to make up various kinds of stresses. At least one of the stressors should be specified. The following is supported stressors for now:

  1. `memory`

     A `memory` stressor will continuously stress virtual memory out. 

     | Option    | Type    | Required | Description                                                  |
     | --------- | ------- | -------- | ------------------------------------------------------------ |
     | `workers` | Integer | True     | Specifies concurrent stressing instance                      |
     | `size`   | String  | False    | Specifies memory size consumed per worker, default is the total available memory. One can also specify the size as *%* of total available memory or in units of *B, KB/KiB, MB/MiB, GB/GiB, TB/TiB*. |

  2. `cpu`

     A `cpu` stressor will continuously stress CPU out. 

     | Option    | Type    | Required | Description                                                  |
     | --------- | ------- | -------- | ------------------------------------------------------------ |
     | `workers` | Integer | True     | Specifies concurrent stressing instance. Actually it specifies how many CPUs to stress when it's less than available CPUs. |
     | `load`    | Integer | False    | Specifies  percent loading per worker. 0 is effectively a sleep (no load) and 100 is full loading |

* `stressngStressors`

  StressngStressors define plenty of stressors just like `Stressors` except that it's an experimental feature and more powerful. You can define stressors in `stress-ng` (see also `man stress-ng`) dialect, however not all of the supported stressors are well tested (**You have been warned**). It may be retired in later releases. You should always use `Stressors` to define the stressors and use this only when you want more stressors unsupported by `Stressors`. When both `StressngStressors` and `Stressors` are defined, `StressngStressors` wins. 

Let's try it out! An example `yaml` of `StressChaos` which burns 1 CPU for 30 seconds in every 2 minutes is listed below: 

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
  stressors:
    cpu:
      workers: 1
  duration: "30s"
  scheduler:
    cron: "@every 2m"
```

Then we could apply it 

```bash
# If you do not have a namespace for your application, you can create one such as tidb-cluster-demo
kubectl create ns tidb-cluster-demo
# Create your pods in the target namespace such as tidb-cluster-demo
kubectl apply -f your-pods.yaml
# Inject a stress chaos
kubectl apply -f your-stress-chaos.yaml
# Your pod's cpu will burn for 30s
```

Have fun : )

