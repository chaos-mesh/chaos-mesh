# Deploy Chaos Mesh

This document describes how to deploy Chaos Mesh to perform chaos experiments against your application in Kubernetes.

## Prerequisites

Before deploying Chaos Mesh, make sure the following items have been installed:

- Kubernetes version >= 1.12
- [RBAC](https://kubernetes.io/docs/admin/authorization/rbac) enabled (optional)
- [Helm](https://helm.sh/) version >= 2.8.2

## Step 1: Get Chaos Mesh

```bash
git clone https://github.com/pingcap/chaos-mesh.git
cd chaos-mesh/
```

## Step 2: Create custom resource type

To use Chaos Mesh, you must create the related custom resource type first.

```bash
kubectl apply -f manifests/crd.yaml
```

## Step 3: Install Chaos Mesh

Depending on your environment, there are two methods of installing Chaos Mesh:

* Install in Docker environment

    1. Create namespace `chaos-testing`

        ```bash
        kubectl create ns chaos-testing
        ```

    2. Install Chaos Mesh using Helm

        * for Helm 2.X

        ```bash
        helm install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing
        ```

        * for Helm 3.X

        ```bash
        helm install chaos-mesh helm/chaos-mesh --namespace=chaos-testing
        ```

    3. Check whether Chaos Mesh pods are installed

        ```
        kubectl get pods --namespace chaos-testing -l app.kubernetes.io/instance=chaos-mesh
        ```

* Install in containerd environment (kind)

    1. Create namespace `chaos-testing`

        ```bash
        kubectl create ns chaos-testing
        ```

    2. Install Chaos Mesh using Helm

        * for Helm 2.X

        ```bash
        helm install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing --set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock
        ```

        * for Helm 3.X

        ```bash
        helm install chaos-mesh helm/chaos-mesh --namespace=chaos-testing --set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock
        ```

    3. Check whether Chaos Mesh pods are installed

        ```bash
        kubectl get pods --namespace chaos-testing -l app.kubernetes.io/instance=chaos-mesh
        ```

> **Note:**
>
> Currently, Chaos Dashboard is not installed by default. If you want to try it out, add `--set dashboard.create=true` in the Helm commands above. Refer to [Configuration](../helm/chaos-mesh/README.md#parameters) for more information.

After executing the above commands, you should be able to see the output indicating that all Chaos Mesh pods are up and running. Otherwise, check the current environment according to the prompt message or create an [issue](https://github.com/pingcap/chaos-mesh/issues) for help.

## Next steps

[Run Chaos Mesh](run_chaos_mesh.md).
