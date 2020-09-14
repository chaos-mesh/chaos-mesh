---
id: sidecar_template
title: Sidecar Template 
sidebar_label: Sidecar Template
---

The following content is the common template ConfigMap defined for injecting IOChaos sidecar, you can also find this example [here](https://github.com/chaos-mesh/chaos-mesh/blob/master/manifests/chaosfs-sidecar.yaml):

## Template ConfigMap

```yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: chaosfs-sidecar
  labels:
    app.kubernetes.io/component: template
data:
  data: |
    initContainers:
    - name: inject-scripts
      image: pingcap/chaos-scripts:latest
      imagePullPolicy: Always
      command: ["sh", "-c", "/scripts/init.sh -d {{.DataPath}} -f {{.MountPath}}/fuse-data"]
    containers:
    - name: chaosfs
      image: pingcap/chaos-fs:latest
      imagePullPolicy: Always
      ports:
      - containerPort: 65533
      securityContext:
        privileged: true
      command:
        - /usr/local/bin/chaosfs
        - -addr=:65533
        - -pidfile=/tmp/fuse/pid
        - -original={{.MountPath}}/fuse-data
        - -mountpoint={{.DataPath}}
      volumeMounts:
        - name: {{.VolumeName}}
          mountPath: {{.MountPath}}
          mountPropagation: Bidirectional
    volumeMounts:
    - name: {{.VolumeName}}
      mountPath: {{.MountPath}}
      mountPropagation: HostToContainer
    - name: scripts
      mountPath: /tmp/scripts
    - name: fuse
      mountPath: /tmp/fuse
    volumes:
    - name: scripts
      emptyDir: {}
    - name: fuse
      emptyDir: {}
    postStart:
      {{.ContainerName}}:
        command:
          - /tmp/scripts/wait-fuse.sh
```

Template config defines some variables by [Go Template](https://golang.org/pkg/text/template/) mechanism. This example has four arguments:

- DataPath: original data directory
- MountPath: after injecting chaosfs sidecar, data directory will be mounted to {{.MountPath}}/fuse-data
- VolumeName: the data volume name used by the pod
- ContainerName: to which container the sidecar is injected

For fields defined in this template, we have some brief descriptions below:

* **initContainers**: defines the [initContainer](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/) need to be injected.
* **container**: defines the sidecar container need to be injected.
* **volumeMounts**: defines the new volumeMounts or overwrite the old volumeMounts of each containers in target pods.
* **volume**: defines the new volumes for the target pod or overwrite the old volumes in target pods.
* **postStart**: called after a container is created first. If the handler fails, the containers will failed.

> **Note:**
>
> Chaos controller-manager only watches template config map with the label selector specified by its flag `--template-labels`, by default this label 
> is `app.kubernetes.io/component=template` if your Chaos Mesh is deployed by helm.
>
> Each template config map should be deployed in the same namespace as Chaos Mesh, and it is identified by the name of the config map, which is `chaosfs-sidecar` in the above example.
>
> The template config content should be in the `data` field. This means it is not possible to define two templates in one config map, you have to use two config maps like the example below.

```yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: chaosfs-sidecar0
  labels:
    app.kubernetes.io/component: template
data:
  data: |
    xxxx

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: chaosfs-sidecar1
  labels:
    app.kubernetes.io/component: template
data:
  data: |
    xxxx
```

### Containers

#### `chaosfs`

`chaosfs` container is designed as a sidecar container and [chaosfs](https://github.com/chaos-mesh/chaos-mesh/tree/master/cmd/chaosfs) process runs in this container.

`chaosfs` uses [fuse libary](https://github.com/hanwen/go-fuse) and [fusermount](https://www.kernel.org/doc/Documentation/filesystems/fuse.txt) tool to implement a fuse-daemon service and mounts the application's data directory. `chaosfs` hijacks all the file system IO actions of the application, so it can be used to simulate various real-world IO faults.

The following configuration injects `chaosfs` container to the target pods and starts a `chaosfs` process in this container.

In addition, `chaosfs` container should be run as `privileged` and the [`mountPropagation`](https://kubernetes.io/docs/concepts/storage/volumes/#mount-propagation) field in `chaosfs` Container.volumeMounts should be set to `Bidirectional`.

`chaosfs` uses `fusermount` to mount the data directory of the application container in `chaosfs` container.

If any Pod with `Bidirectional` mount propagation to the same volume mounts anything there, the Container with `HostToContainer` mount propagation will see it.

This mode is equal to `rslave` mount propagation as described in the [Linux kernel documentation](https://www.kernel.org/doc/Documentation/filesystems/sharedsubtree.txt).

More detail about `Mount propagation` can be found [here](https://kubernetes.io/docs/concepts/storage/volumes/#mount-propagation).

```yaml
containers:
- name: chaosfs
  image: pingcap/chaos-fs
  imagePullpolicy: Always
  ports:
  - containerPort: 65534
  securityContext:
    privileged: true
  command:
    - /usr/local/bin/chaosfs
    - -addr=:65534
    - -pidfile=/tmp/fuse/pid
    - -original=/var/lib/tikv/fuse-data
    - -mountpoint=/var/lib/tikv/data
  volumeMounts:
    - name: tikv
      mountPath: /var/lib/tikv
      mountPropagation: Bidirectional
```

Description of `chaosfs`:

* **addr**: defines the address of the grpc server, default value: ":65534".
* **pidfile**: defines the pid file to record the pid of the `chaosfs` process.
* **original**: defines the fuse directory. This directory is usually set to the same level directory as the application data directory.
* **mountpoint**: defines the mountpoint to mount the original directory.

This value should be set to the data directory of the target application.

#### `chaos-scripts`

`chaos-scripts` container is used to inject some scripts to the target pods including [wait-fuse.sh](https://github.com/chaos-mesh/chaos-mesh/blob/master/scripts/wait-fuse.sh).

`wait-fuse.sh` is used by application container to ensure that the fuse-daemon server is running normally before the application starts.

`chaos-scripts` is generally used as an initContainer to do some preparations.

The following config uses `chaos-scripts` container to inject scripts and moves the scripts to `/tmp/scripts` directory using `init.sh`. `/tmp/scripts` is an [emptyDir volume](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) used to share the scripts with all containers of the pod.

So you can use `wait-fuse.sh` script in tikv container to ensure that the fuse-daemon server is running normally before the application starts.

In addition, `init.sh` creates a directory named `fuse-data` in the [PersistentVolumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) directory of the tikv as the original directory for fuse-daemon server and the original directory is required.

You should also create the original directory in the PersistentVolumes directory of the application.

```yaml
initContainers:
  - name: inject-scripts
    image: pingcap/chaos-scripts:latest
    imagePullpolicy: Always
    command: ["sh", "-c", "/scripts/init.sh -d /var/lib/tikv/data -f /var/lib/tikv/fuse-data"]
```

The usage of `init.sh`:

```bash
$ ./scripts/init.sh -h
```

Expected output:

```bash
USAGE: ./scripts/init.sh [-d data directory] [-f fuse directory]
Used to do some preparation
OPTIONS:
   -h                      Show this message
   -d <data directory>     Data directory of the application
   -f <fuse directory>     Data directory of the fuse original directory
   -s <scripts directory>  Scripts directory
EXAMPLES:
   init.sh -d /var/lib/tikv/data -f /var/lib/tikv/fuse-data
```

### Tips

1. The application Container.volumeMounts used to define data directory should be set to `HostToContainer`.
2. `scripts` and `fuse` emptyDir should be created and should be mounted to all container of the pod.
3. The application uses `wait-fuse.sh` script to ensure that the fuse-daemon server is running normally.

```yaml
postStart:
  tikv:
    command:
      - /tmp/scripts/wait-fuse.sh
```

The usage of `wait-fuse.sh`:

```bash
$ ./scripts/wait-fuse.sh -h
```

Expected output:

```bash
./scripts/wait-fuse.sh: option requires an argument -- h
USAGE: ./scripts/wait-fuse.sh [-a <host>] [-p <port>]
Waiting for fuse server ready
OPTIONS:
   -h                   Show this message
   -f <host>            Set the target file
   -d <delay>           Set the delay time
   -r <retry>           Set the retry count
EXAMPLES:
   wait-fuse.sh -f /tmp/fuse/pid -d 5 -r 60
```
