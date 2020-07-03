---
id: installation
title: Installation
---

This document describes how to install Chaos Mesh to perform chaos experiments against your application in Kubernetes.

If you want to try Chaos Mesh on your your laptop (Linux or macOS), you can refer the following two documents:

- [Get started on kind](get_started_on_kind.md)
- [Get started on minikube](get_started_on_minikube.md)

## Prerequisites

Before deploying Chaos Mesh, make sure the following items have been installed:

- Kubernetes version >= 1.12
- [RBAC](https://kubernetes.io/docs/admin/authorization/rbac) enabled (optional)

## Install Chaos Mesh

```bash
curl -sSL https://raw.githubusercontent.com/pingcap/chaos-mesh/master/install.sh | bash
```

The above command install all the CRDs, required service account configuration, and all components.
Before you start running a chaos experiment, verify if Chaos Mesh is installed correctly.

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
curl -sSL https://raw.githubusercontent.com/pingcap/chaos-mesh/master/install.sh | sh -s -- --template | kubectl delete -f -
```

## Install by Helm

You also can install Chaos Mesh by [Helm](https://helm.sh).
Before you start installing, make sure that Helm v2 or Helm v3 is installed correctly.

### Step 1: Get Chaos Mesh

```bash
git clone https://github.com/pingcap/chaos-mesh.git
cd chaos-mesh/
```

### Step 2: Create custom resource type

To use Chaos Mesh, you must create the related custom resource type first.

```bash
kubectl apply -f manifests/crd.yaml
```

### Step 3: Install Chaos Mesh

Depending on your environment, there are two methods of installing Chaos Mesh:

- Install in Docker environment

  1. Create namespace `chaos-testing`:

     ```bash
     kubectl create ns chaos-testing
     ```

  2. Install Chaos Mesh using Helm:

     - For Helm 2.X

     ```bash
     helm install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing
     ```

     - For Helm 3.X

     ```bash
     helm install chaos-mesh helm/chaos-mesh --namespace=chaos-testing
     ```

  3. Check whether Chaos Mesh pods are installed:

     ```bash
     kubectl get pods --namespace chaos-testing -l app.kubernetes.io/instance=chaos-mesh
     ```

- Install in containerd environment (kind)

  1. Create namespace `chaos-testing`:

     ```bash
     kubectl create ns chaos-testing
     ```

  2. Install Chaos Mesh using Helm:

     - for Helm 2.X

     ```bash
     helm install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing --set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock
     ```

     - for Helm 3.X

     ```bash
     helm install chaos-mesh helm/chaos-mesh --namespace=chaos-testing --set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock
     ```

  3. Check whether Chaos Mesh pods are installed:

     ```bash
     kubectl get pods --namespace chaos-testing -l app.kubernetes.io/instance=chaos-mesh
     ```

> **Note:**
>
> Currently, Chaos Dashboard is not installed by default. If you want to try it out, add `--set dashboard.create=true` in the Helm commands above. Refer to [Configuration](../helm/chaos-mesh/README.md#configuration) for more information.

After executing the above commands, you should be able to see the output indicating that all Chaos Mesh pods are up and running. Otherwise, check the current environment according to the prompt message or create an [issue](https://github.com/pingcap/chaos-mesh/issues) for help.
