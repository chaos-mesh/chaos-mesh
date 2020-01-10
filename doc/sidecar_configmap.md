# Sidecar ConfigMap

This document guides user to define a specify sidecar ConfigMap for your application.

## Why need a specify Sidecar ConfigMap?

Chaos Mesh runs a [fuse-daemon](https://www.kernel.org/doc/Documentation/filesystems/fuse.txt) server in [sidecar container](https://www.magalix.com/blog/the-sidecar-pattern) for implementing file system IO Chaos. 
In sidecar container, fuse-daemon need to mount the data directory of application by [fusermount](http://manpages.ubuntu.com/manpages/bionic/en/man1/fusermount.1.html) before the application starts.
The most applications use different data directories, so we need to defines the different sidecar configs for most different applications. 

## What in Sidecar ConfigMap?

The following content is the sidecar ConfigMap defined for tikv: 

```yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: chaosfs-tikv
  labels:
    app.kubernetes.io/component: webhook
data:
  # the real content of config
  chaosfs-tikv.yaml: |
    name: chaosfs-tikv
    selector:
      labelSelectors:
        "app.kubernetes.io/component": "tikv"
    initContainers:
    - name: inject-scripts
      image: pingcap/chaos-scripts:latest
      imagePullpolicy: Always
      command: ["sh", "-c", "mkdir -p /tmp/scripts; cp -R /scripts/* /tmp/scripts/; mkdir -p /var/lib/tikv/fuse-data"]
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
        - -original=/var/lib/tikv/data
        - -mountpoint=/var/lib/tikv/fuse-data
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
> For more sample ConfigMap files, see [examples](https://github.com/pingcap/chaos-mesh/tree/master/examples/chaosfs-configmap)

Description of `chaofs-tikv.yaml`: 

* **name**: defines the name of the sidecar config, this name should be unique across all sidecar configs.
* **selector**: is used to filter pods to inject sidecar.
* **initContainers**: defines the [initContainer](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/) need to be injected.
* **container**: defines the sidecar container need to be injected. 
* **volumeMounts**: defines the new volumeMounts or overwrite the old volumeMounts of the each containers in target pods.
* **volume**: defines the new volumes for the target pod or overwrite the old volumes in target pods.
* **postStart**: called after a container is created first. If the handler fails, the containers will failed. 
Key defines for the name of deployment container. Value defines for the Commands for stating container.

### Containers

#### chaosfs

`chaosfs` container is designed as a sidecar container and [chaosfs](https://github.com/pingcap/chaos-mesh/tree/master/cmd/chaosfs) process runs in this container.
`chaosfs` uses [fuse libary](https://github.com/hanwen/go-fuse) and [fusermount](https://www.kernel.org/doc/Documentation/filesystems/fuse.txt) tool to implement a fuse-daemon service 
and mounts the application's data directory. `chaosfs` will hijack all the file system IO cations of the application, so it can be used to simulate various IO fault that often occur in the real world. 

The following config will inject `chaosfs` container to target pods and will start a `chaosfs` process in this container.
In addition, `chaosfs` container should be run as `privileged` and the [mountPropagation](https://kubernetes.io/docs/concepts/storage/volumes/#mount-propagation) 
field in `chaosfs` Container.volumeMounts should be set to `Bidirectional`.
`chaosfs` will use `fusermount` to mount the data directory of the application container in `chaosfs` container. 
If any Pod with `Bidirectional` mount propagation to the same volume mounts anything there, the Container with `HostToContainer` mount propagation will see it.
This mode is equal to `rslave` mount propagation as described in the [Linux kernel documentation](https://www.kernel.org/doc/Documentation/filesystems/sharedsubtree.txt).

> More detail about `Mount propagation` can be found [here](https://kubernetes.io/docs/concepts/storage/volumes/#mount-propagation)

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
    - -original=/var/lib/tikv/data
    - -mountpoint=/var/lib/tikv/fuse-data
  volumeMounts:
    - name: tikv
      mountPath: /var/lib/tikv
      mountPropagation: Bidirectional
```

Description of `chaosfs`:

* **addr**: defines the address of the grpc server, default value: ":65534".
* **pidfile**: defines the pid file to record the pid of the `chaosfs` process.
* **original**: defines the target directory that need to be injected file system IO fault.
This value should be set to the data directory of the target application.
* **mountpoint**: defines the mountpoint to mount original directory.
This directory is usually set to the same level directory as the original directory.

#### chaos-scripts

`chaos-scripts` container is used to inject some scripts to target pods, include [wait-fuse.sh](https://github.com/pingcap/chaos-mesh/blob/master/hack/wait-fuse.sh).
`wait-fuse.sh` is used by application container to ensure that the fuse-daemon server is running normally before the application starts. 

`chaos-scripts` is generally used as an initContainer to do some preparation. 
The following config uses `chaos-scripts` container to inject scripts and moves the scripts to `/tmp/scripts` directory, 
`/tmp/scripts` is an [emptyDir volume](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) to shares the scripts with all containers of the pod.
So you can use `wait-fuse.sh` script in tikv container to ensure that the fuse-daemon server is running normally before the application starts.   

In addition, this config created a directory named `fuse-data` in the [PersistentVolumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) directory of the tikv 
as the [mountpoint](https://www.kernel.org/doc/Documentation/filesystems/fuse.txt) for fuse-daemon server and the mountpoint directory is required.
You should also create the mountpoint directory in the PersistentVolumes directory of the application.

```yaml
initContainers:
  - name: inject-scripts
    image: pingcap/chaos-scripts:latest
    imagePullpolicy: Always
    command: ["sh", "-c", "mkdir -p /tmp/scripts; cp -R /scripts/* /tmp/scripts/; mkdir -p /var/lib/tikv/fuse-data"]
```

### Tips

1. The application Container.volumeMounts used to define data directory should be set `HostToContainer`.
2. `scripts` and `fuse` emptyDir should be created and should be mounted to all container of the pod.
3. The application uses `wait-fuse.sh` script to ensure that the fuse-daemon server is running normally.

```yaml
postStart:
  tikv:
    command:
      - /tmp/scripts/wait-fuse.sh
```

## How to use?

You can apply the ConfigMap defined for your application to Kubernetes cluster by using this command:

```bash
kubectl apply -f app-configmap.yaml # app-configmap.yaml is the ConfigMap file 
```

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

If the target application is a TiDB cluster, you can follow the instructions in the following two documents to deploy one:

* [Deploy using kind](https://pingcap.com/docs/stable/tidb-in-kubernetes/get-started/deploy-tidb-from-kubernetes-kind/)
* [Deoloy using minikube](https://pingcap.com/docs/stable/tidb-in-kubernetes/get-started/deploy-tidb-from-kubernetes-minikube/)


Then, you can start your application and define your [IO Chaos](io_chaos.md) config to start your chaos experiment.

