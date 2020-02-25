# Get started on minikube

This document describes how to deploy Chaos Mesh in Kubernetes on your laptop (Linux or macOS) using minikube.

## Prerequisites

Before deployment, make sure [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) 
is installed on your local machine.

## Step 1: Setup the Kubernetes environment

Take the following steps to set up the local Kubernetes environment:

1. Start a `minikube` kubernetes cluster. Make sure you have installed [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/).

   ```bash
   minikube start --kubernetes-version v1.15.0 --cpus 4 --memory "8192mb" # we recommend that you allocate enough RAM (more than 8192 MiB) to the VM
   ```

2. Install Helm

   ```bash
   curl https://raw.githubusercontent.com/helm/helm/master/scripts/get | bash
   helm init
   ```

3. Check whether the Helm tiller pod is running.

   ```bash
   kubectl -n kube-system get pods -l app=helm
   ```

## Step 2: Install Chaos Mesh

Run the following comments to install Chaos Mesh;

```bash
git clone --depth=1 https://github.com/pingcap/chaos-mesh && \
cd chaos-mesh
./install.sh
```

>**Note:** 
>
> `install.sh` is a shell script to automate the installation process. To deploy manually,  refer to [Install Chaos Mesh manually](deploy.md).

> **Limitations:**
>
> There are some known restrictions for Chaos Operator deployed on `minikube` clusters:
> - `netem chaos` is only supported for `minikube` clusters >= version 1.6.
>
>   In `minikube`, the default virtual machine driver's image doesn't contain the `sch_netem` kernel module in smaller versions. You can use `none` driver (if your host is Linux with the `sch_netem` kernel module loaded) to try these chaos actions on `minikube` or [build an image with sch_netem by yourself](https://minikube.sigs.k8s.io/docs/contributing/iso/).

## Step 3: Run Chaos Mesh

Refer to the Steps in [Run Chaos Mesh](run-chaos-mesh.md)