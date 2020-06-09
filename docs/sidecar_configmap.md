# Sidecar ConfigMap

This document guides you to define a specified sidecar ConfigMap for your application.

## Why do we need a specified Sidecar ConfigMap?

Chaos Mesh runs a [fuse-daemon](https://www.kernel.org/doc/Documentation/filesystems/fuse.txt) server in [sidecar container](https://www.magalix.com/blog/the-sidecar-pattern) for implementing file system IO Chaos.

In sidecar container, fuse-daemon needs to mount the data directory of application by [fusermount](http://manpages.ubuntu.com/manpages/bionic/en/man1/fusermount.1.html) before the application starts.

## How it works?

Currently, Chaos Mesh supports two types of ConfigMaps:

1. Template config. The skeleton of each sidecar config is similar, in order to fulfill different requirements and make the configuration simplified,
Chaos Mesh supports creating common templates to be used by different applications. For the details of template configuration, please refer to [template config](./template_config.md).

2. Injection config. This configuration will be combined with template config and finally generate a config to inject to the selected pods. 
Since most applications use different data directories, volume name or container name, you can define different parameters based on the common template created in the first step.

## Injection Configuration

The following content is an injection ConfigMap defined for tikv:

```yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: chaosfs-tikv
  namespace: chaos-testing
  labels:
    app.kubernetes.io/component: webhook
data:
  chaosfs-tikv: |
    name: chaosfs-tikv
    selector:
      labelSelectors:
        "app.kubernetes.io/component": "tikv"
    template: chaosfs-sidecar
    arguments:
      ContainerName: "tikv"
      DataPath: "/var/lib/tikv/data"
      MountPath: "/var/lib/tikv"
      VolumeName: "tikv"
```

Injection config defines some injection arguments for different applications, and it is based on the common template created beforehand.

For fields defined in this config, we have some brief descriptions below:

* **name**: injection config name, uniquely identifies a injection config in one namespace. 
  However, you can have the same name in different namespaces so this is useful to implement multi-tenancy.
* **selector**: is used to filter pods to inject sidecar.
* **template**: the template config map name used to render the injection config.
* **arguments**: the arguments you should define to be used in the template.

The final injection config content is rendered by `template` and `arguments` via `Go Template` and 
will be injected to the selected pods. In this example, the final injection config is:

```
    name: chaosfs-tikv
    selector:
      labelSelectors:
        "app.kubernetes.io/component": "tikv"    
    initContainers:
    - name: inject-scripts
      image: pingcap/chaos-scripts:latest
      imagePullPolicy: Always
      command: ["sh", "-c", "/scripts/init.sh -d /var/lib/tikv/data -f /var/lib/tikv/fuse-data"]
    containers:
    - name: chaosfs
      image: pingcap/chaos-fs:latest
      imagePullPolicy: Always
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
    volumeMounts:
    - name: tikv
      mountPath: /var/lib/tikv
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
      tikv:
        command:
          - /tmp/scripts/wait-fuse.sh
```

For more sample ConfigMap files, see [examples](../examples/chaosfs-configmap).

### Containers

#### `chaosfs`

`chaosfs` container is designed as a sidecar container and [chaosfs](https://github.com/pingcap/chaos-mesh/tree/master/cmd/chaosfs) process runs in this container.

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

`chaos-scripts` container is used to inject some scripts to the target pods including [wait-fuse.sh](https://github.com/pingcap/chaos-mesh/blob/master/hack/wait-fuse.sh).

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

## Usage

See [IOChaos Document](io_chaos.md).
