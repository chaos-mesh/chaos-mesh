---
id: get_started_on_minikube
title: Get started on Minikube
---

This document describes how to deploy Chaos Mesh in Kubernetes on your laptop (Linux or macOS) using Minikube.

## Prerequisites

Before deployment, make sure [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) is installed on your local machine.

## Setp 1: Set up the Kubernetes environment

Perform the following steps to set up the local Kubernetes environment:

1. Start a Kubernetes cluster:

   ```bash
   minikube start --kubernetes-version v1.15.0 --cpus 4 --memory "8192mb"
   ```

    > **Note:**
    >
    > It is recommended to allocate enough RAM (more than 8192 MiB) to the Virtual Machine (VM) using the `--cpus` and `--memory` flag.

2. Install helm:

   ```bash
   curl https://raw.githubusercontent.com/helm/helm/master/scripts/get | bash
   helm init
   ```

3. Check whether the helm tiller pod is running:

   ```bash
   kubectl -n kube-system get pods -l app=helm
   ```

## Setp 2: Install Chaos Mesh

```bash
curl -sSL https://raw.githubusercontent.com/chaos-mesh/chaos-mesh/master/install.sh | bash
```

The above command install all the CRDs, required service account configuration, and all components.
Before you start running a chaos experiment, verify if Chaos Mesh is installed correctly.

You also can use [helm](https://helm.sh/) to [install Chaos Mesh manually](installation.md#install-by-helm).

### Verify your installation

Verify if the chaos mesh is running

```bash
kubectl get pod -n chaos-testing
```

Expected output: 

```bash
NAME                                        READY   STATUS    RESTARTS   AGE
chaos-controller-manager-6d6d95cd94-kl8gs   1/1     Running   0          3m40s
chaos-daemon-5shkv                          1/1     Running   0          3m40s
chaos-daemon-jpqhd                          1/1     Running   0          3m40s
chaos-daemon-n6mfq                          1/1     Running   0          3m40s
chaos-dashboard-d998856f6-vgrjs             1/1     Running   0          3m40s
```

## Uninstallation

You can uninstall Chaos Mesh by deleting the namespace.

```bash
curl -sSL https://raw.githubusercontent.com/chaos-mesh/chaos-mesh/master/install.sh | bash -s -- --template | kubectl delete -f -
```

## Limitations

There are some known restrictions for Chaos Operator deployed in the Minikube cluster:

- `netem chaos` is only supported for Minikube clusters >= version 1.6.

In Minikube, the default virtual machine driver's image does not contain the `sch_netem` kernel module in earlier versions. You can use `none` driver (if your host is Linux with the `sch_netem` kernel module loaded) to try these chaos actions using Minikube or [build an image with sch_netem by yourself](https://minikube.sigs.k8s.io/docs/contrib/building/iso/).
