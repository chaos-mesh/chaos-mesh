# Time Chaos Document

This document describe how to add time chaos experiments in Chaos Mesh.

Time chaos is used to modify the return value of `clock_gettime`, which will lead to time offset on go's `time.Now()`, rust std's 'std::time::Instant::now()' etc.

Below is a sample time chaos configuration file:

```yaml
apiVersion: pingcap.com/v1alpha1
kind: NetworkChaos
metadata:
  name: network-partition-example
  namespace: chaos-testing
spec:
  mode: one
  selector:
    namespaces:
      - tidb-cluster-demo
    labelSelectors:
      "app.kubernetes.io/component": "pd"
  timeOffset:
    sec: 100000
    nsec: 100000
  clockIds:
    - CLOCK_REALTIME
  duration: "10s"
  scheduler:
    cron: "@every 15s"
```

> For more sample files, see [examples](../examples). You can edit them as needed. 

Description:

* **mode** defines the mode to select pods.
* **selector** specifies the target pods for chaos injection.
* **timeOffset** specifies the offset of time. `sec` means the offset of seconds and `nsec` means the offset of nanoseconds.
* **clockIds** defines all affected `clk_id`. `clk_id` refers to the first argument of `clock_gettime` call. For most application, `CLOCK_REALTIME` is enough.
* **duration** defines the duration for each chaos experiment. In the sample file above, the time chaos lasts for 10 seconds.
* **scheduler** defines the scheduler rules for the running time of the chaos experiment. For more rule information, see <https://godoc.org/github.com/robfig/cron>.

## Limitation

* Time modification will only be injected into the main process of container
* Time chaos has no effect on pure syscall `clock_gettime`
* All injected vdso call will use pure syscall to get realtime, so clock related function call will be much slower.