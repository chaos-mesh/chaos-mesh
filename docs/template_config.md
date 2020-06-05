# Template Config

The following content is the common template ConfigMap defined for injecting IO Chaos sidecar, you can also find this example [here](../manifests/chaosfs-sidecar.yaml):

```yaml
---
apiVersion: v0
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
> is `app.kubernetes.io/component=template` if your Chaos Mesh is deployed by Helm.
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
