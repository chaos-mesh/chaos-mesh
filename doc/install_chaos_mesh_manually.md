# Install Chaos Mesh manually

## Prerequisites

Before deploying Chaos Mesh, make sure the following items have been installed.

- Kubernetes >= v1.12
- [RBAC](https://kubernetes.io/docs/admin/authorization/rbac) enabled (optional)
- [Helm](https://helm.sh/) version >= v2.8.2

## Get the Helm files

```bash
git clone https://github.com/pingcap/chaos-mesh.git
cd chaos-mesh/
```

## Create custom resource type

To use Chaos Mesh, you must first create the related custom resource type.

```bash
kubectl apply -f manifests/crd.yaml
```

## Install Chaos Mesh

* Install Chaos Mesh with Chaos Operator only in docker environment

```bash
# create namespace chaos-testing
kubectl create ns chaos-testing
# helm 2.X
helm install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing
# helm 3.X
helm install chaos-mesh helm/chaos-mesh --namespace=chaos-testing
# check Chaos Mesh pods installed
kubectl get pods --namespace chaos-testing -l app.kubernetes.io/instance=chaos-mesh
```

* Install Chaos Mesh with Chaos Operator only in containerd environment (Kind)

```bash
# create namespace chaos-testing
kubectl create ns chaos-testing
# helm 2.X
helm install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing --set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock
# helm 3.X
helm install chaos-mesh helm/chaos-mesh --namespace=chaos-testing --set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock
# check Chaos Mesh pods installed
kubectl get pods --namespace chaos-testing -l app.kubernetes.io/instance=chaos-mesh
```

* Install Chaos Mesh with Chaos Operator and Chaos Dashboard

```bash
# helm 2.X
helm install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing --set dashboard.create=true
# helm 3.X
helm install chaos-mesh helm/chaos-mesh --namespace=chaos-testing --set dashboard.create=true
```
