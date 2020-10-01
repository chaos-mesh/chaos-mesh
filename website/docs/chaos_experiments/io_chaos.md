---
id: iochaos_experiment
title: IOChaos Experiment
sidebar_label: IOChaos Experiment
---

This document walks you through the IOChaos experiment.

IOChaos allows you to simulate file system faults such as IO delay and read/write errors. It can inject delay and fault when your program is running IO system calls such as `open`, `read`, and `write`.

## Configuration file

Below is a sample YAML file of IOChaos:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IoChaos
metadata:
  name: io-delay-example
spec:
  action: latency
  mode: one
  selector:
    labelSelectors:
      app: etcd
  volumePath: /var/run/etcd
  path: "/var/run/etcd/**/*"
  delay: "100ms"
  percent: 50
  duration: "400s"
  scheduler:
    cron: "@every 10m"
```

For more sample files, see [examples](https://github.com/chaos-mesh/chaos-mesh/tree/master/examples). You can edit them as needed.

| Field | Description | Sample Value |
|:------|:------------------|:--------------|
| **mode** | Defines the mode of the selector. | `one` / `all` / `fixed` / `fixed-percent` / `random-max-percent` |
| **selector** | Specifies the pods to be injected with IO chaos. |
| **action** | Represents the IOChaos actions. Refer to [Available actions for IOChaos](#iavailable-actions-for-iochaos) for more details. | `delay` / `fault` / `attrOverride` |
| **volumePath** | The mount path of the target volume | `"/var/run/etcd"` |
| **delay** | Specifies the latency of the fault injection. The duration might be a string with a signed sequence of decimal numbers, each with an optional fraction and a unit suffix. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", and "h". | `"300ms"` / `"2h45m"` |
| **errno** | Defines the error code returned by an IO action. See [common Linux system errors](#common-linux-system-errors) for more Linux system error codes. | `2` |
| **attr** | Defines the attribute to be overridden and the corresponding value | [examples](https://github.com/chaos-mesh/chaos-mesh/tree/master/examples/io-attr-example.yaml) |
| **percent** | Defines the probability of injecting errors in percentage. | `100` (by default) |
| **path** | Defines the path of files for injecting IOChaos actions. It should be a glob for the files which you want to inject fault or delay. | "/var/run/etcd/*\*/\*" |
| **methods** | Defines the IO methods for injecting IOChaos actions. It is represented as an array of string. | `open` / `read` See the [available methods](#available-methods) for more details. |
| **duration** | Represents the duration of a chaos action. The duration might be a string with the signed sequence of decimal numbers, each with an optional fraction and a unit suffix. | `"300ms"` / `"2h45m"`|
| **scheduler** | Defines the scheduler rules for the running time of the chaos experiment. | see [robfig/cron](https://godoc.org/github.com/robfig/cron) |

## Usage

Assume that you are using `examples/io-mixed-example.yaml`, you can run the following command to create a chaos experiment:

```bash
kubectl apply -f examples/io-mixed-example.yaml
```

## IOChaos available actions

IOChaos currently supports the following actions:

* **delay**: IO delay action. You can specify the latency before the IO operation returns a result.
* **fault**: IO fault action. In this mode, IO operations returns an error.
* **attrOverride**: Override attributes of a file.

### delay

If you are using the `delay` action, you can edit the specification as below:

```yaml
spec:
  action: delay
  delay: "1ms"
```

It will inject a latency of 1ms into the selected methods.

### fault

If you are using the `fault` action, you can edit the specification  as below:

```yaml
spec:
  action: fault
  errno: 32
```

The selected methods return error 32, which means `broken pipe`.

### attrOverride

If you are using the `attrOverride` mode, you can edit the specification as below:

```yaml
spec:
  action: attrOverride
  attr:
    perm: 72
```

Then the permission of selected files will be overridden with 110 in octal, which means the files cannot be read or modified (without CAP_DAC_OVERRIDE). See [available attributes](#available-attributes) for a list of all possible attributes to override.

> **Note:
>
> Attributes could be cached by Linux kernel, so it might have no effect if your program had accessed it before.**

## Common Linux system errors

Common Linux system errors are as below:

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

Refer to [related header files](https://raw.githubusercontent.com/torvalds/linux/master/include/uapi/asm-generic/errno-base.h) for more information.

## Available methods

Available methods are as below:

* lookup
* forget
* getattr
* setattr
* readlink
* mknod
* mkdir
* unlink
* rmdir
* symlink
* rename
* link
* open
* read
* write
* flush
* release
* fsync
* opendir
* readdir
* releasedir
* fsyncdir
* statfs
* setxattr
* getxattr
* listxattr
* removexattr
* access
* create
* getlk
* setlk
* bmap

## Available attributes

Available attributes and the meaning of them are listed here:

* `ino`, inode of a file
* `size`, total size, in bytes
* `blocks`, number of 512B blocks allocated
* `atime`, time of last access
* `mtime`, time of last modification
* `ctime`, time of last status change
* `kind`, file type. It can be `namedPipe`, `charDevice`, `blockDevice`, `directory`, `regularFile`, `symlink` or `socket`
* `perm`, permission of a file
* `nlink`, number of hard links
* `uid`, user id of owner
* `gid`, group id of owner
* `rdev`, device ID (if special file)
