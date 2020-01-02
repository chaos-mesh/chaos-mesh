# IO Chaos document

This document will help user to build IO Chaos experiments. 

IO Chaos can help user simulate file system faults such as I/O delay, read/write errors, etc. It can inject delay and errno when user using syscall about IO like `open`, `read`, `write`. 

## Sample Config

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

User can find and edit the template refer to [examples/io-mixed-sample.yaml](../examples/io-mixed-sample.yaml).

## Usage

### Config

We use [Kubernetes annotations](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/) to  attach metadata about IO chaos to objects. In [examples/io-mixed-sample.yaml](../examples/io-mixed-sample.yaml), user can find metadata below.

```yaml
metadata:
  name: io-delay-example
  namespace: chaos-testing
```

### Run

Assuming user are using `examples/io-mixed-sample.yaml`, to create a chaos experiment:

```bash
kubectl apply -f examples/io-mixed-sample.yaml
```

## Spec Arguements

* **selector**: is used to select pods that are used to inject chaos action.

* **action**: action represents the chaos action about IO action, now the **delay**, **errno**,  **mixed** action is supported. User can go to [*IO Chaos Availiable Actions*](#io-chaos-availiable-actions) for more details.
* **mode**: Mode defines the mode to run chaos action. Supported mode: `one` / `all` / `fixed` / `fixed-percent` / `random-max-percent`.
* **duration**: represents the duration of the chaos action. The duration is a possibly string with signed sequence of decimal numbers,  each with optional fraction and a unit suffix, such as `"300ms"`, `"-1.5h"` or `”2h45m"`.
* **delay**: defines the value of I/O chaos action delay. The duration is a possibly string with signed sequence of decimal numbers,  each with optional fraction and a unit suffix, such as `"300ms"`, `"-1.5h"` or `”2h45m”`. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
  If `Delay` is empty, the operator will generate a value for it randomly.
* **errno**: defines the error code that returned by IO action. It is a int32 string like `"32"`. This field should be set when user choose `errno`  or `mixed` action. If `errno` is empty, the operator will generate a error code for it randomly. User can set the `errno` refer to: https://www-numi.fnal.gov/offline_software/srt_public_context/WebDocs/Errors/unix_system_errors.html.
* **percent**: Percent defines the percentage of injection errors and provides a number from 0-100. The defualt value is `100`.
* **path**: defines the path of files for injecting I/O chaos action. It should be an regular expression for the path user want to inject errno or delay. If path is `""` or not defined, IO to all files will be injected.
* **methods**: defines the I/O methods for injecting I/O chaos action. It’s an array of string, which set the IO syscall like `open` `read`. User can see the [availiable methods](#availiable-methods) below.
* **addr**: defines the sidecar HTTP server address for sidecar container, like `":8080"`.
* **configName**: defines the config name which used to inject pod. User can refer to [examples/tikv-configmap.yaml](../../examples/tikv-configmap.yaml) to define user's config.
* **layer**: represents the layer of the I/O action. Supported value: `fs` , and default is `fs`.

## IO Chaos Availiable Actions

IO Chaos now support the actions below:

* **delay**: IO delay action. User can specify the latency before the IO operation will return.
* **errno**: IO errno action. In this mode read/write IO operation will return error.IO errno means user's read/write IO operations will return error.
* **mixed**: Both **delay** and **errno** actions.

### delay

If user are using delay mode, user may edit spec like:

```yaml
spec:
  action: delay
  delay: "1ms"
```

If `delay` is not specified, it will be generate randomly on runtime.

### errno

```yaml
spec:
  action: errno
  errno: "32"
```

If `errno` is not specified, it will be generate randomly on runtime. 

### mixed

````yaml
spec:
  action: mixed
  delay: "1ms"
  errno: "32"
````

It is mix of **delay** and **errno**.

## Availiable Methods

Availiable methods are:

* `open`
* `read`
* `write`
* `mkdir`
* `rmdir`
* `opendir`
* `fsync`
* `flush`
* `release`
* `truncate`
* `getattr`
* `chown`
* `chmod`
* `utimens`
* `allocate`
* `getlk`
* `setlk`
* `setlkw`
* `statfs`
* `readlink`
* `symlink`
* `create`
* `access`
* `link`
* `mknod`
* `rename`
* `unlink`
* `getxattr`
* `listxattr`
* `removexattr`
* `setxattr`