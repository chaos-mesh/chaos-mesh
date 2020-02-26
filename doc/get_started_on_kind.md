# Get started on kind

This document describes how to deploy Chaos Mesh in Kubernetes on your laptop (Linux or macOS) using kind.

## Prerequisites

Before deployment, make sure [Docker](https://docs.docker.com/install/) is installed and running on your local machine.

## Step 1: Get Chaos Mesh

```bash
git clone https://github.com/pingcap/chaos-mesh.git
cd chaos-mesh/
```

## Step 2: Install Chaos Mesh

```bash
./install.sh --local kind
```

`install.sh` is an automation shell script that helps you install dependencies such as `kubelet`, `Helm`, `kind`, and `kubernetes`, and `Chaos Mesh` itself.

After executing the above command, you should be able to see the prompt that Chaos Mesh is installed successfully. 
Otherwise, please check the current environment according to the prompt message or send us an [issue](https://github.com/pingcap/chaos-mesh/issues) for help. 
In addition, You also can use [Helm](https://helm.sh/) to [install Chaos Mesh manually](deploy.md).


## Next steps

Refer to [Run Chaos Mesh](run_chaos_mesh.md).

