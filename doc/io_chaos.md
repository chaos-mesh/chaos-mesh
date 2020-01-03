# IO Chaos Document

This document helps you to build IO chaos experiments. 

IO chaos allows you to simulate file system faults such as IO delay, read/write errors, etc. It can inject delay and errno when you use the IO system calls such as `open`, `read` and `write`.

## Sample config file

Here is a sample YAML file of IO chaos:

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

For more sample files, see [examples/io-mixed-example.yaml](../examples/io-mixed-example.yaml). You can edit them as needed. 

## Usage

### Configuration

#### Annotations

We use [Kubernetes annotations](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/) to attach IO chaos metadata to objects. You should set annotations for namespace and name. In [examples/io-mixed-example.yaml](../examples/io-mixed-example.yaml), you can find the metadata as below.

```yaml
metadata:
  name: io-delay-example
  namespace: chaos-testing
```

Note that if you do not attach an annotation to the namespace,  the pod will be modified dynamically, and might be restarted.

#### Data directory

The data directory of the component of the pod should be a subdirectory of `PersistentVolumes`.

#### Addmission-webhook

You should make sure admission-webhooks is turned on, refer: https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#experimenting-with-admission-webhooks .

### Create a chaos experiment

Assume that you are using `examples/io-mixed-example.yaml`, you can run the following command to create a chaos experiment:

```bash
kubectl apply -f examples/io-mixed-example.yaml
```

## Spec arguments

* **selector**: is used to select pods that are used to inject chaos actions.

* **action**: represents the IO chaos actions. Currently the **delay**, **errno**, and **mixed** actions are supported. You can go to [*IO chaos available actions*](#io-chaos-available-actions) for more details.
* **mode**: defines the mode to run chaos actions. Supported mode: `one` / `all` / `fixed` / `fixed-percent` / `random-max-percent`.
* **duration**: represents the duration of a chaos action. The duration might be a string with the signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as `"300ms"`, `"-1.5h"` or `"2h45m"`.
* **delay**: defines the value of IO chaos action delay. The duration might be a string with the signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as `"300ms"`, `"-1.5h"` or `”2h45m”`. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", and "h".
  If `Delay` is empty, the operator will generate a value for it randomly.
* **errno**: defines the error code that is returned by an IO action. It is an int32 string like `"32"`. This field need to be set when you choose an `errno` or `mixed` action. If `errno` is empty, the operator will randomly generate an error code for it. You can set the `errno` by referring to [Errors: Linux System Errors](https://www-numi.fnal.gov/offline_software/srt_public_context/WebDocs/Errors/unix_system_errors.html).
* **percent**: defines the percentage of injection errors and provides a number from 0-100. The default value is `100`.
* **path**: defines the path of files for injecting IO chaos actions. It should be a regular expression for the path you want to inject errno or delay. If the path is `""` or not defined, IO chaos actions will be injected into all files.
* **methods**: defines the IO methods for injecting IO chaos actions. It’s an array of string, which sets the IO syscalls such as `open` and `read`. See the [available methods](#available-methods) for more details.
* **addr**: defines the sidecar HTTP server address for a sidecar container, such as `":8080"`.
* **configName**: defines the config name which is used to inject chaos action into pods. You can refer to [examples/tikv-configmap.yaml](../examples/tikv-configmap.yaml) to define your configuration.
* **layer**: represents the layer of the IO action. Supported value: `fs` (by default).

## IO chaos available actions

IO chaos currently supports the following actions:

* **delay**: IO delay action. You can specify the latency before the IO operation returns a result.
* **errno**: IO errno action. In this mode, read/write IO operations will return an error.
* **mixed**: Both **delay** and **errno** actions.

### delay

If you are using the delay mode, you can edit spec as below:

```yaml
spec:
  action: delay
  delay: "1ms"
```

If `delay` is not specified, it will be generated randomly on runtime.

### errno

If you are using the errno mode, you can edit spec as below:

```yaml
spec:
  action: errno
  errno: "32"
```

If `errno` is not specified, it will be generated randomly on runtime. 

### mixed

If you are using the mixed mode, you can edit spec as below:

````yaml
spec:
  action: mixed
  delay: "1ms"
  errno: "32"
````

The mix mode defines the **delay** and **errno** actions in one spec.

## Available methods

Available methods are as below:

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
