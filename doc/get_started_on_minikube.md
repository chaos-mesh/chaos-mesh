# Get Started on Minikube

This document describes how to deploy Chaos Mesh in Kubernetes on your laptop (Linux or macOS) using Minikube.

## Prerequisites

Before deployment, make sure [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) is installed on your local machine.

## Step 1: Set up the Kubernetes environment

Perform the following steps to set up the local Kubernetes environment:

1. Start a Kubernetes cluster:

   ```bash
   minikube start --kubernetes-version v1.15.0 --cpus 4 --memory "8192mb"
   ```

    > **Note:**
    >
    > It is recommended to allocate enough RAM (more than 8192 MiB) to the Virtual Machine (VM) using the `--cpus` and `--memory` flag.

2. Install Helm:

   ```bash
   curl https://raw.githubusercontent.com/helm/helm/master/scripts/get | bash
   helm init
   ```

3. Check whether the Helm tiller pod is running:

   ```bash
   kubectl -n kube-system get pods -l app=helm
   ```

## Step 2: Install Chaos Mesh

Run the following comments to install Chaos Mesh:

```bash
git clone --depth=1 https://github.com/pingcap/chaos-mesh && \
cd chaos-mesh
./install.sh
```

>**Note:**
>
> `install.sh` is a shell script to automate the installation process. See [Install Chaos Mesh manually](deploy.md) for more details.

After executing the above commands, you should be able to see the prompt that Chaos Mesh is installed successfully. Otherwise, check the current environment according to the prompt message or send us an [issue](https://github.com/pingcap/chaos-mesh/issues) for help.

## Limitations

There are some known restrictions for Chaos Operator deployed in the Minikube cluster:

- `netem chaos` is only supported for Minikube clusters >= version 1.6.

In Minikube, the default virtual machine driver's image doesn't contain the `sch_netem` kernel module in earlier versions. You can use `none` driver (if your host is Linux with the `sch_netem` kernel module loaded) to try these chaos actions using Minikube or [build an image with sch_netem by yourself](https://minikube.sigs.k8s.io/docs/contributing/iso/).

## Next steps

[Run Chaos Mesh](run_chaos_mesh.md).
