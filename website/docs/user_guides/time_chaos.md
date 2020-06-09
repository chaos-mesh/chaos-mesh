---
id: timechaos_experiment
title: TimeChaos Experiment
sidebar_label: TimeChaos Experiment
---

This document describe how to add TimeChaos experiments in Chaos Mesh.

TimeChaos is used to modify the return value of `clock_gettime`, which causes time offset on Go's `time.Now()` and Rust std's `std::time::Instant::now()` etc.

## Configuration file

Below is a sample TimeChaos configuration file:

```yaml
apiVersion: pingcap.com/v1alpha1
kind: TimeChaos
metadata:
  name: time-shift-example
  namespace: chaos-testing
spec:
  mode: one
  selector:
    labelSelectors:
      "app.kubernetes.io/component": "pd"
  timeOffset: "-10m100ns"
  clockIds:
    - CLOCK_REALTIME
  containerNames:
    - pd
  duration: "10s"
  scheduler:
    cron: "@every 15s"
```

For more sample files, see [examples](https://github.com/pingcap/chaos-mesh/tree/master/examples). You can edit them as needed.

Description:

* **mode** defines the mode to select pods.
* **selector** specifies the target pods for chaos injection.
* **timeOffset** specifies the time offset. It is a duration string with specified unit, such as `300ms`, `-1.5h`. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
* **clockIds** defines all affected `clk_id`. `clk_id` refers to the first argument of `clock_gettime` call. For most application, `CLOCK_REALTIME` is enough.
* **containerNames** selects affected containers' names. If not set, all containers will be injected.
* **duration** defines the duration for each chaos experiment. In the sample file above, the time chaos lasts for 10 seconds.
* **scheduler** defines the scheduler rules for the running time of the chaos experiment. For more rule information, see <https://godoc.org/github.com/robfig/cron>.

## Limitation

* Time modification can only be injected into the main process of container.
* Time chaos has no effect on pure system call `clock_gettime`.
* All injected [vDSO](http://man7.org/linux/man-pages/man7/vdso.7.html) calls use pure system calls to get the real time, so clock-related function calls can be much slower.
