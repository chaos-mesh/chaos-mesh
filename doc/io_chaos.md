# IO Chaos Document

This document helps you build IO chaos experiments.

IO chaos allows you to simulate file system faults such as IO delay and read/write errors. It can inject delay and errno when you use IO system calls such as `open`, `read` and `write`.

> **Note:**
>
> IO Chaos can only be used if the relevant labels and annotations are set before the application is created. See [Create a chaos experiment](#create-a-chaos-experiment) for more information.

## Prerequisites

### Commands and arguments for the application container

Chaos Mesh uses [`wait-fush.sh`](https://github.com/pingcap/chaos-mesh/blob/master/doc/sidecar_configmap.md#tips) to ensure that the fuse-daemon server is running normally before the application starts.

Therefore, `wait-fush.sh` needs to be injected into the startup command of the container. If the application process is not started by the [commands and arguments of the container](https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/), IO chaos cannot work properly.

>**Note:**
>
> When Kubernetes natively supports [Sidecar Containers](https://github.com/kubernetes/enhancements/issues/753) in future versions, we will remove the `wait-fush.sh` dependency.

### Admission Controller

IO chaos needs to inject a sidecar container to user pods and the sidecar container can be added to applicable Kubernetes pods using a [mutating webhook admission controller](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/) provided by Chaos Mesh.

> **Note:**
>
> * While admission controllers are enabled by default, some Kubernetes distributions may disable them. In this case, follow the instructions to [turn on admission controllers](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#how-do-i-turn-on-an-admission-controller).
> * [ValidatingAdmissionWebhooks](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#validatingadmissionwebhook) and [MutatingAdmissionWebhooks](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook) are required by IO chaos.

### Data directory

The data directory of the application in the target pod should be a **subdirectory** of `PersistentVolumes`.

Example:

```yaml
# the config about tikv PersistentVolumes
volumeMounts:
  - name: datadir
    mountPath: /var/lib/tikv

# the arguments to start tikv
ARGS="--pd=${CLUSTER_NAME}-pd:2379 \
  --advertise-addr=${HOSTNAME}.${HEADLESS_SERVICE_NAME}.${NAMESPACE}.svc:20160 \
  --addr=0.0.0.0:20160 \
  --data-dir=/var/lib/tikv/data \  # data directory
  --capacity=${CAPACITY} \
  --config=/etc/tikv/tikv.toml
```

> **Note:**
>
> * The default data directory of TiKV is not a subdirectory of `PersistentVolumes`.
> * If you are testing a TiDB cluster, you need to modify it at [`_start_tikv.sh.tpl`](https://github.com/pingcap/tidb-operator/blob/master/charts/tidb-cluster/templates/scripts/_start_tikv.sh.tpl).
> * PD has the same issue with TiKV. You need to modify the data directory of PD at [`_start_pd.sh.tpl`](https://github.com/pingcap/tidb-operator/blob/master/charts/tidb-cluster/templates/scripts/_start_pd.sh.tpl).

## Usage

### Configure a ConfigMap

Chaos Mesh uses sidecar container to inject IO chaos. To fulfill this chaos, you need to configure this sidecar container using a [ConfigMap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/).

To define a specified ConfigMap for your application before starting your chaos experiment, refer to this [document](sidecar_configmap.md).

You can apply the ConfigMap defined for your application to Kubernetes cluster by the following command:

```bash
kubectl apply -f app-configmap.yaml # app-configmap.yaml is the ConfigMap file
```

### Define the configuration file

Below is a sample YAML file of IO Chaos:

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
    labelSelectors:
      "app.kubernetes.io/component": "tikv"
  layer: "fs"
  percent: "50"
  delay: "1ms"
  scheduler:
    cron: "@every 10m"
```

For more sample files, see [examples](../examples). You can edit them as needed.

| Field | Description | Sample Value |
|:------|:------------------|:--------------|
| **selector** | Selects pods that are used to inject chaos actions.|
| **action** | Represents the IO chaos actions.| `delay` / `errno` / `mixed`. Refer to [IO chaos available actions](#io-chaos-available-actions) for more details.|
| **mode** | Defines the mode to run chaos actions.| `one` / `all` / `fixed` / `fixed-percent` / `random-max-percent`.|
| **duration** | Represents the duration of a chaos action. The duration might be a string with the signed sequence of decimal numbers, each with optional fraction and a unit suffix. | `"300ms"`, `"-1.5h"` or `"2h45m"`.|
| **delay** | Defines the value of IO chaos action delay. The duration might be a string with the signed sequence of decimal numbers, each with optional fraction and a unit suffix. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", and "h". If `Delay` is empty, the operator will generate a value for it randomly.| `"300ms"`, `"-1.5h"` or `"2h45m"`. |
| **errno** | Defines the error code that is returned by an IO action. This value and the [errno defined by Linux system](http://man7.org/linux/man-pages/man3/errno.3.html) are consistent. This field needs to be set when you choose an `errno` or `mixed` action. If `errno` is empty, the operator randomly generates an error code for it. See the [common Linux system errors](#common-linux-system-errors) for more Linux system error codes. | An int32 string like `"2"`, which means `No such file or directory`. |
| **percent** | Defines the percentage of injection errors and provides a number from 0-100.| `100` (by default) |
| **path** | Defines the path of files for injecting IO chaos actions. It should be a regular expression for the path you want to inject errno or delay. If the path is `""` or not defined, IO chaos actions will be injected into all files.| |
| **methods** | Defines the IO methods for injecting IO chaos actions. It’s an array of string, which sets the IO syscalls. See the [available methods](#available-methods) for more details.| `open` and `read` |
| **addr** | Defines the sidecar HTTP server address for a sidecar container.| `":8080"` |
| **configName** | Defines the config name which is used to inject chaos action into pods. You can refer to [examples/tikv-configmap.yaml](../examples/chaosfs-configmap/tikv-configmap.yaml) to define your configuration.| |
| **layer** | Represents the layer of the IO action.| `fs` (by default). |

### Create a chaos experiment

Before the application created, you need to make admission-webhook enabled using labels and [annotations](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/) to the application namespace:

```bash
admission-webhook.pingcap.com/init-request:chaosfs-tikv
```

You can use the following commands to set labels and annotations of the application namespace:

```bash
# If the application namespace does not exist. you can exec this command to create one,
# otherwise ignore this command.
kubectl create ns app-ns # "app-ns" is the application namespace

# enable admission-webhook
kubectl label ns app-ns admission-webhook=enabled

# set annotation
kubectl annotate ns app-ns admission-webhook.pingcap.com/init-request=chaosfs-tikv

# create your application
...
```

Then, you can start your application and define YAML file to start your chaos experiment.

#### Start a chaos experiment

Assume that you are using `examples/io-mixed-example.yaml`, you can run the following command to create a chaos experiment:

```bash
kubectl apply -f examples/io-mixed-example.yaml
```

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

If `delay` is not specified, it is generated randomly on runtime.

### errno

If you are using the errno mode, you can edit spec as below:

```yaml
spec:
  action: errno
  errno: "32"
```

If `errno` is not specified, it is generated randomly on runtime.

### mixed

If you are using the mixed mode, you can edit spec as below:

````yaml
spec:
  action: mixed
  delay: "1ms"
  errno: "32"
````

The mix mode defines the **delay** and **errno** actions in one spec.

## Common Linux system errors

The number represents the errno the Linux system error.

* `1`: Operation not permitted
* `2`: No such file or directory
* `5`: I/O error
* `6`: No such device or address
* `12`: Out of memory
* `16`: Device or resource busy
* `17`: File exists
* `20`: Not a directory
* `22`: Invalid argument
* `24`: Too many open files
* `28`: No space left on device

For more Linux system errors, refer to [Errors: Linux System Errors](https://www-numi.fnal.gov/offline_software/srt_public_context/WebDocs/Errors/unix_system_errors.html).

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
