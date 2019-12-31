# IO Chaos document

Sample IO chaos ducument:

```yaml
apiVersion: pingcap.com/v1alpha1
kind: IoChaos
metadata:
  name: io-delay-example
  namespace: chaos-testing
spec:
  action: mixed
  mode: one
  duration: "400s"
  configName: "chaosfs-tikv"
  path: ""
  selector:
    namespaces:
      - tidb-cluster-demo
    labelSelectors:
      "app.kubernetes.io/component": "tikv"
  layer: "fs"
  percent: "50"
  delay: "1ms"
  scheduler:
    cron: "@every 10m"	
```

The file can be find in [examples](../examples).

## Spec Arguements

* **action**: action represents the chaos action about IO action, now the following action is supported:
* **delay**: IO delay action. In this mode read/write IO operation will return error.
  * **errno**: IO errno action.You can specify the latency beore the IO operation will return. IO errno means your read/write IO operation will return error.
  * **mixed**: Both delay and errno actions.
* **mode**: Mode defines the mode to run chaos action. Supported mode: `one` / `all` / `fixed` / `fixed-percent` / `random-max-percent`.
* **duration**: represents the duration of the chaos action. The duration is a possibly string with signed sequence of decimal numbers,  each with optional fraction and a unit suffix, such as `"300ms"`, `"-1.5h"` or `”2h45m"`.
* **errno**: defines the error code that returned by IO action. This field should be set when you choose `errno`  or `mixed` action. If `Errno` is empty, the operator will generate a error code for it randomly. You can set the `Errno` refer to: https://www-numi.fnal.gov/offline_software/srt_public_context/WebDocs/Errors/unix_system_errors.html.
* **percent**: Percent defines the percentage of injection errors and provides a number from 0-100. The defualt value is `100`.
* **path**: defines the path of files for injecting I/O chaos action. It should be an regular expression for the path user want to inject errno or delay. If path is `""` or not defined, IO to all files will be injected.
* **methods**: defines the I/O methods for injecting I/O chaos action. It’s an array of string, which set the IO syscall like `open` `read` .
* **addr**: defines the address for sidecar container.
* **configName**: defines the config name which used to inject pod.
* **nextStart**: Next time when this action will be applied again.
* **nextRecover**: Next time when this action will be recovered.

