# Kernel Chaos Document

This document describe how to add kernel chaos experiments in Chaos Mesh.

Kernel chaos's function is similar to
[inject.py](https://github.com/iovisor/bcc/blob/master/tools/inject.py), which
guarantees the appropriate erroneous return of the specified injection mode
(kmalloc,bio,etc) given a call chain and an optional set of predicates.

You can read
[inject\_example.txt](https://github.com/iovisor/bcc/blob/master/tools/inject_example.txt)
to learn more.

Below is a sample kernel chaos configuration file:

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

And a sample program:
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

During the injection, the program will output:

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

> For more sample files, see [examples](https://github.com/chaos-mesh/bpfki/tree/develop/examples). You can edit them as needed. 

Description:

* **mode** defines the mode to select pods.
* **selector** specifies the target pods for chaos injection.
* **failkernRequest** defines the specified injection mode (kmalloc,bio,etc) with a call chain and an optional set of predicates.
* **duration** defines the duration for each chaos experiment. In the sample file above, the time chaos lasts for 10 seconds.
* **scheduler** defines the scheduler rules for the running time of the chaos experiment. For more rule information, see <https://godoc.org/github.com/robfig/cron>.

## Limitation

* Although we use container\_id to limit fault injection, but some behaviors may
  trigger systemic behaviors. For example, when failtype is 1, it means that
  physical page allocation will fail, and if the behavior is continuous in a 
  very short time (eg: ``while (1) {memset(malloc(1M), '1', 1M)}`), the system's
  oom-killer will be awakened to release memory. So the container\_id will lose
  limit to oom-killer. In this case, you may try it with `probability` and `times`.
