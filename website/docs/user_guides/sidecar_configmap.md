---
id: sidecar_configmap
title: Sidecar ConfigMap 
sidebar_label: Sidecar ConfigMap 
---

This document guides you to define a specified sidecar ConfigMap for your application.

## Why do we need a specified Sidecar ConfigMap?

Chaos Mesh runs a [fuse-daemon](https://www.kernel.org/doc/Documentation/filesystems/fuse.txt) server in [sidecar container](https://www.magalix.com/blog/the-sidecar-pattern) for implementing file system IOChaos.

In sidecar container, fuse-daemon needs to mount the data directory of application by [fusermount](http://manpages.ubuntu.com/manpages/bionic/en/man1/fusermount.1.html) before the application starts.

## How it works?

Currently, Chaos Mesh supports two types of ConfigMaps:

1. Template config. The skeleton of each sidecar config is similar, in order to fulfill different requirements and make the configuration simplified,
Chaos Mesh supports creating common templates to be used by different applications. For the details of template configuration, please refer to [template config](sidecar_template.md).

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
* **template**: the template config map name used to render the injection config. "chaosfs-sidecar" template is used for injecting fuse-server sidecar.
* **arguments**: the arguments you should define to be used in the template.

For more sample ConfigMap files, see [examples](https://github.com/pingcap/chaos-mesh/tree/master/examples/chaosfs-configmap).


## Usage

See [IOChaos Document](io_chaos.md).


