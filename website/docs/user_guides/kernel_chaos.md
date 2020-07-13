---
id: kernelchaos_experiment
title: KernelChaos Experiment
sidebar_label: KernelChaos Experiment
---

This document describes how to create KernelChaos experiments in Chaos Mesh.

Although KernelChaos targets a certain pod, the performance of other pods are also impacted depending on the specific callchain and frequency. It is because all pods of the same host share the same kernel.

> **Warning:**
>
> This feature is disabled by default. Do not use it in production environment.

## Prerequisites

- Linux kernel: version >= 4.18
- [CONFIG_BPF_KPROBE_OVERRIDE](https://cateee.net/lkddb/web-lkddb/BPF_KPROBE_OVERRIDE.html) enabled
- `bpfki.create = true` in [values.yaml](https://github.com/pingcap/chaos-mesh/blob/master/helm/chaos-mesh/values.yaml)

## Configuration file

Below is a sample KernelChaos configuration file:

```yaml
apiVersion: pingcap.com/v1alpha1
kind: KernelChaos
metadata:
  name: kernel-chaos-example
  namespace: chaos-testing
spec:
  mode: one
  selector:
    namespaces:
      - chaos-mount
  failKernRequest:
    callchain:
        - funcname: "__x64_sys_mount"
    failtype: 0
```

For more sample files, see [examples](https://github.com/pingcap/chaos-mesh/tree/master/examples). You can edit them as needed.

Description:

* **mode** defines the mode to select pods.
* **selector** specifies the target pods for chaos injection. For more details, see [Define the Scope of Chaos Experiment](experiment_scope.md).
* **failkernRequest** defines the specified injection mode (kmalloc, bio, etc.) with a call chain and an optional set of predicates. The fields are:
  * **failtype** indicates what to fail, can be set to `0` / `1` / `2`.
    - If `0`, indicates slab to fail (should_failslab)
    - If `1`, indicates alloc_page to fail (should_fail_alloc_page)
    - If `2`, indicates bio to fail (should_fail_bio)

    For more information, see [fault-injection](https://www.kernel.org/doc/html/latest/fault-injection/fault-injection.html) and [inject_example](http://github.com/iovisor/bcc/blob/master/tools/inject_example.txt).

  * **callchain** indicates a special call chain, such as:

       ```c
     ext4_mount
       -> mount_subtree
          -> ...
             -> should_failslab
       ```

      With an optional set of predicates and an optional set of parameters, which used with predicates. See [call chain and predicate examples](https://github.com/chaos-mesh/bpfki/tree/develop/examples) to learn more. If there is no special call chain, just keep `callchain` empty, which means it will fail at any call chain with slab alloc (eg: kmalloc).

      The challchain's type is an array of frames, the frame has three fields:

      * **funcname** can be find from kernel source or `/proc/kallsyms`, such as `ext4_mount`.
      * **parameters** is used with predicate, for example, if you want to inject slab error in `d_alloc_parallel(struct dentry *parent, const struct qstr *name)` with a special name `bananas`, you need to set it to `struct dentry *parent, const struct qstr *name`otherwise omit it.
      * **predicate** accesses the arguments of this frame, example with parameters's, you can set it to `STRNCMP(name->name, "bananas", 8)` to make inject only with it, or omit it to inject for all d_alloc_parallel call chain.
  * **headers** indicates the appropriate kernel headers you need. Eg: "linux/mmzone.h", "linux/blkdev.h" and so on.
  * **probability** indicates the fails with probability. If you want 1%, please set this field with `1`.
  * **times** indicates the max times of fails.
* **duration** defines the duration for each chaos experiment. In the sample file above, the time chaos lasts for 10 seconds.
* **scheduler** defines the scheduler rules for the running time of the chaos experiment. For more rule information, see <https://godoc.org/github.com/robfig/cron>

## Usage

KernelChaos's function is similar to [inject.py](https://github.com/iovisor/bcc/blob/master/tools/inject.py), which guarantees the appropriate erroneous return of the specified injection mode (kmalloc, bio, etc.) given a call chain and an optional set of predicates.

You can read [inject\_example.txt](https://github.com/iovisor/bcc/blob/master/tools/inject_example.txt) to learn more.

Below is a sample program:

```c
#include <sys/mount.h>
#include <stdio.h>
#include <string.h>
#include <errno.h>
#include <unistd.h>

int main(void) {
	int ret;
	while (1) {
		ret = mount("/dev/sdc", "/mnt", "ext4",
			    MS_MGC_VAL | MS_RDONLY | MS_NOSUID, "");
		if (ret < 0)
			fprintf(stderr, "%s\n", strerror(errno));
		sleep(1);
		ret = umount("/mnt");
		if (ret < 0)
			fprintf(stderr, "%s\n", strerror(errno));
	}
}
```

During the injection, the output is similar to this:

```
> Cannot allocate memory
> Invalid argument
> Cannot allocate memory
> Invalid argument
> Cannot allocate memory
> Invalid argument
> Cannot allocate memory
> Invalid argument
> Cannot allocate memory
> Invalid argument
```

## Limitation

Although we use container_id to limit fault injection, but some behaviors might trigger systemic behaviors. For example:

When `failtype` is `1`, it means that physical page allocation will fail. If the behavior is continuous in a very short time (eg: ``while (1) {memset(malloc(1M), '1', 1M)}`), the system's oom-killer will be awakened to release memory. So the container_id will lose limit to oom-killer.
