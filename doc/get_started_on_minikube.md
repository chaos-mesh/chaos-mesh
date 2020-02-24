# Getting started on minikube

## Deploy 

### Prerequisites

* [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)

### Setup the Kubernetes environment
1. Start a `minikube` kubernetes cluster. Make sure you have installed [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/).

   ```bash
   minikube start --kubernetes-version v1.15.0 --cpus 4 --memory "8192mb" # we recommend that you allocate enough RAM (more than 8192 MiB) to the VM
   ```

2. Install helm

   ```bash
   curl https://raw.githubusercontent.com/helm/helm/master/scripts/get | bash
   helm init
   ```

3. Check whether helm tiller pod is running

   ```bash
   kubectl -n kube-system get pods -l app=helm
   ```

### Install `chaos-mesh` 

* Run the following comments to install Chaos Mesh

```bash
git clone --depth=1 https://github.com/pingcap/chaos-mesh && \
cd chaos-mesh
./install.sh
```

* Install Chaos Mesh manually refer to [Install Chaos Mesh manually](deploy.md).

**Note:**

There are some known restrictions for Chaos Operator deployed on `minikube` clusters:

- `netem chaos` is only supported for `minikube` clusters >= version 1.6.

    In `minikube`, the default virtual machine driver's image doesn't contain the `sch_netem` kernel module in smaller versions. You can use `none` driver (if your host is Linux with the `sch_netem` kernel module loaded) to try these chaos actions on `minikube` or [build an image with sch_netem by yourself](https://minikube.sigs.k8s.io/docs/contributing/iso/).

## Usage

Refer to the Steps in [Usage](usage.md)