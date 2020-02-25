# Get started on kind

This document describes how to deploy Chaos Mesh in Kubernetes on your laptop (Linux or macOS) using kind.

## Prerequisites

Before deploying Chaos Mesh, make sure the following items have been installed. 

- [Docker](https://docs.docker.com/install/) (required when running in [kind](https://kind.sigs.k8s.io/))

## Step 1: Get the Helm files

```bash
git clone https://github.com/pingcap/chaos-mesh.git
cd chaos-mesh/
```

## Step 2: Install Chaos Mesh

```bash
./install.sh --local kind
```

`install.sh` will help you to install `kubelet`, `Helm`, `kind`, `kubernetes` and `Chaos Mesh`. 

After executing the above command, if the message that Chaos Mesh is installed 
successfully is output normally, then you can continue next steps to test your application and enjoy Chaos Mesh. 
Otherwise, please check the current environment according to the prompt message of the script 
or send us an [issue](https://github.com/pingcap/chaos-mesh/issues/new/choose) for help. 
In addition, You also can use [Helm](https://helm.sh/) to [install Chaos Mesh manually](deploy.md).

## Step 3: Run Chaos Mesh

Refer to the Steps in [Run Chaos Mesh](run_chaos_mesh.md)

