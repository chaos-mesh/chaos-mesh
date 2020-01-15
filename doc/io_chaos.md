# IO Chaos Document

This document helps you to build IO chaos experiments. 

IO chaos allows you to simulate file system faults such as IO delay, 
read/write errors, etc. It can inject delay and errno when you use the IO system calls such as `open`, `read` and `write`.

> Note: IO Chaos can only be used if the relevant labels and annotations are set before the application is created. 
> More info refer [here](#create-a-chaos-experiment)

## Prerequisites

### Admission Controller

IO chaos needs to inject a sidecar container to user pods and the sidecar container can be added to applicable Kubernetes pods 
using a [mutating webhook admission controller](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/) provided by Chaos Mesh.

> While admission controllers are enabled by default, some Kubernetes distributions may disable them. 
> If this is the case, follow the instructions to [turn on admission controllers](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#how-do-i-turn-on-an-admission-controller).     
> [ValidatingAdmissionWebhooks](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#validatingadmissionwebhook) and [MutatingAdmissionWebhooks](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook) are required by IO chaos.

### Data directory

The data directory of the application in the target pod should be a **subdirectory** of `PersistentVolumes`.

example:
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

> Node: The default data directory of TiKV is not a subdirectory of `PersistentVolumes`.
> If your application is TiDB cluster, you need to modify it at [_start_tikv.sh.tpl](https://github.com/pingcap/tidb-operator/blob/master/charts/tidb-cluster/templates/scripts/_start_tikv.sh.tpl). 
> PD has the same issue with TiKV, you need to modity the data directory of pd at [_start_pd.sh.tpl](https://github.com/pingcap/tidb-operator/blob/master/charts/tidb-cluster/templates/scripts/_start_pd.sh.tpl).

## Usage

### Configure a ConfigMap 

Chaos Mesh uses sidecar container to inject IO chaos, 
to fulfill this chaos you need to configure this sidecar container using a [ConfigMap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/)
You can refer this [document](sidecar_configmap.md) to define a specify ConfigMap for your application before starting your chaos experiment. 

You can apply the ConfigMap defined for your application to Kubernetes cluster by using this command:

```bash
kubectl apply -f app-configmap.yaml # app-configmap.yaml is the ConfigMap file 
```

### Define the Chaos YAML file

Below is a sample YAML file of IO chaos:

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

> For more sample files, see [examples](../examples). You can edit them as needed. 

Description: 

* **selector**: is used to select pods that are used to inject chaos actions.

* **action**: represents the IO chaos actions. Currently the **delay**, **errno**, and **mixed** actions are supported. You can go to [*IO chaos available actions*](#io-chaos-available-actions) for more details.
* **mode**: defines the mode to run chaos actions. Supported mode: `one` / `all` / `fixed` / `fixed-percent` / `random-max-percent`.
* **duration**: represents the duration of a chaos action. The duration might be a string with the signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as `"300ms"`, `"-1.5h"` or `"2h45m"`.
* **delay**: defines the value of IO chaos action delay. The duration might be a string with the signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as `"300ms"`, `"-1.5h"` or `”2h45m”`. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", and "h".
  If `Delay` is empty, the operator will generate a value for it randomly.
* **errno**: defines the error code that is returned by an IO action. It and [errno](http://man7.org/linux/man-pages/man3/errno.3.html) defined by Linux system are consistent. It is an int32 string like `"2"`, `"2"` means `No such file or directory`. 
This field need to be set when you choose an `errno` or `mixed` action. If `errno` is empty, the operator will randomly generate an error code for it. 
See the [common Linux system errors](#common-linux-system-errors) for more Linux system error codes.
* **percent**: defines the percentage of injection errors and provides a number from 0-100. The default value is `100`.
* **path**: defines the path of files for injecting IO chaos actions. It should be a regular expression for the path you want to inject errno or delay. If the path is `""` or not defined, IO chaos actions will be injected into all files.
* **methods**: defines the IO methods for injecting IO chaos actions. It’s an array of string, which sets the IO syscalls such as `open` and `read`. 
See the [available methods](#available-methods) for more details.
* **addr**: defines the sidecar HTTP server address for a sidecar container, such as `":8080"`.
* **configName**: defines the config name which is used to inject chaos action into pods. You can refer to [examples/tikv-configmap.yaml](../examples/chaosfs-configmap/tikv-configmap.yaml) to define your configuration.
* **layer**: represents the layer of the IO action. Supported value: `fs` (by default).

### Create a chaos experiment

Before the application created, you need to make admission-webhook enable by label add an [annotation](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/) to the application namespace:

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

## Common Linux system errors

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

> More Linux system errors refer to [Errors: Linux System Errors](https://www-numi.fnal.gov/offline_software/srt_public_context/WebDocs/Errors/unix_system_errors.html).

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
