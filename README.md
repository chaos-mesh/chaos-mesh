<img src="static/logo.png" alt="chaos_logo" width="450"/>

> **Note:**
>
> This readme and related documentation are a Work in Progress.

Chaos Mesh is a cloud-native Chaos Engineering toolset that orchestrates chaos on Kubernetes environment. At the current stage, it has the following components:

- **Chaos Operator**: the core component for Chaos orchastration. Fully open sourced.
- **Chaos Dashboard**: a visualized panel that shows the impacts of Chaos experiments on the online services of the system; under development; curently only supports chaos experiments on TiDB.

[![Watch the video](./static/demo.gif)](https://www.youtube.com/watch?v=ifZEwdJO868)

## Chaos Operator

Chaos Operator injects chaos into the applications and Kubernetes infrastructure in a manageable way, which provides easy, custom definitions for chaos experiments and automatic orchastration. There are three components at play:

**Controller-manager**: used to schedule and manage the lifecycle of CRD objects

**Chaos-daemon**: runs as daemonset with previleged system permissions over network, Cgroup, etc. on each node

**Sidecar**: a special type of container that is dynamically injected into the target Pod by the webhook-server, which can be used for hacjacking I/O of the application container.

![Chaos Operator](./static/chaos-mesh-overview.png)

Chaos Operator uses [Custom Resource Definition (CRD)](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/) to define chaos objects. The current implementation supports three types of CRD objects for fault injection, namely PodChaos, NetworkChaos, and IOChaos, which correspond to the following major actions (experiments):

- pod-kill: The selected pod is killed (ReplicaSet or something similar may be needed to ensure the pod will be restarted)
- pod-failure: The selected pod will be unavailable in a specified period of time
- netem chaos: Network chaos such as delay, duplication, etc.
- network-partition: Simulate network partition
- IO chaos: simulate file system falults such as I/O delay, read/write errors, etc.

## Prerequisites

Before deploying Chaos Mesh, make sure the following items have been installed. If you would like to have a try on your machine, you can refer to [get-started-on-your-local-machine](#get-started-on-your-local-machine) section.

* Kubernetes >= v1.12 and < v1.16
* [RBAC](https://kubernetes.io/docs/admin/authorization/rbac) enabled (optional)
* [Helm](https://helm.sh/) version >= v2.8.2 and < v3.0.0

## Deploy Chaos Mesh

### Get the Helm files

```bash
git clone https://github.com/pingcap/chaos-mesh.git
cd chaos-mesh/
```

### Create custom resource type

To use Chaos Mesh, you must first create the related custom resource type.

```bash
kubectl apply -f manifests/
kubectl get crd podchaos.pingcap.com
```

### Install Chaos Mesh

* Install Chaos Mesh with Chaos Operator only

```bash
helm install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing
kubectl get pods --namespace chaos-testing -l app.kubernetes.io/instance=chaos-mesh
```

* Install Chaos Mesh with Chaos Operator and Chaos Dashboard

```bash
helm install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing --set dashboard.create=true
```

## Get started on your local machine

> **Warning:**
>
>**This deployment is for testing only. DO NOT USE in production!**

You can try Chaos Mesh on your local K8s environment deployed using `kind` or `minikube`.

### Deploy your local K8s environment

#### Deploy with `kind`

1. Clone the code

   ```bash
   git clone --depth=1 https://github.com/pingcap/chaos-mesh && \
   cd chaos-mesh
   ```

2. Run the script and create a local Kubernetes cluster

   ```bash
   hack/kind-cluster-build.sh
   ```

3. To connect the local Kubernetes cluster, set the default configuration file path of `kubectl` to `kube-config`.

   ```bash
   export KUBECONFIG="$(kind get kubeconfig-path)"
   ```

4. Verify whether the Kubernetes cluster is on and running

   ```bash
   kubectl cluster-info
   ```

5. Install `chaos-mesh` on `kind` kubernetes cluster as suggested in [Deploy Chaos Mesh](#deploy-chaos-mesh).

#### Deploy with `minikube`

1. Start a `minikube` kubernetes cluster

   ```bash
   minikube start --kubernetes-version v1.15.0 --cpus 4 --memory "8192mb" # we recommend that you allocate enough RAM(better more than 8192 MiB) to VM
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

4. Install `chaos-operator` as suggested in [Deploy Chaos Mesh](#deploy-chaos-mesh).

**Note:**

There are some known restrictions for Chaos Operator deployed on `kind` and `minikube` clusters:

- All network related chaos is not supported for `Kind` cluster.

     Chaos Operator uses docker pkg to transform between container id and pid, which is necessary to find network namespace for pods.`Kind` uses `containerd` as Introducing Container Runtime Interface (CRI) runtime and it's not supported in our implementation yet.

- `netem chaos` is not supported for `minikube` clusters.

    In `minikube`, the default virtual machine driver's image doesn't contain the `sch_netem` kernel module. You can use `none` driver (if your host is Linux with the `sch_netem` kernel module loaded) to try these chaos actions on `minikube` or [build a image with sch_netem by yourself](https://minikube.sigs.k8s.io/docs/contributing/iso/).

### Deploy target cluster

After Chaos Mesh is deployed, we can deploy the target cluster to be tested, or where we want to inject faults. For illustration purposes, we use TiDB as our sample cluster.

You can follow the instructions on the following two document to deploy a TiDB cluster:

* [if use kind](https://pingcap.com/docs/stable/tidb-in-kubernetes/get-started/deploy-tidb-from-kubernetes-kind/)
* [if use minikube](https://pingcap.com/docs/stable/tidb-in-kubernetes/get-started/deploy-tidb-from-kubernetes-minikube/)

### Define chaos experiment config file

In this sample experiment config file, we will define a chaos experiment to kill one tikv pod randomly:

```yaml
apiVersion: pingcap.com/v1alpha1
kind: PodChaos
metadata:
  name: pod-failure-example
  namespace: chaos-testing
spec:
  action: pod-kill # the specific chaos action to inject; supported action: pod-kill/pod-failure
  mode: one # the mode to run chaos action; supported mode are one/all/fixed/fixed-percent/random-max-percent
  duration: "60s" # duration for the injected chaos experiment
  selector: # pods where to inject chaos actions
    namespaces:
      - tidb-cluster-demo
    labelSelectors:
      "app.kubernetes.io/component": "tikv"
  scheduler: #defines scheduler rules for the running time of the chaos experiments about pods.
    cron: "@every 5m"
```

### Create a chaos experiment

```bash
kubectl apply -f pod-failure-example.yaml
kubectl get podchaos --namespace=chaos-testing
```

You can see the QPS performance affected by the chaos experiment from TiDB Grafana dashboard:

![tikv-pod-failure](./static/tikv-pod-failure.png)

### Update a chaos experiment

```bash
vim pod-failure-example.yaml # modify pod-failure-example.yaml to what you want
kubectl apply -f pod-failure-example.yaml
```

#### Delete a chaos experiment

```bash
kubectl delete -f pod-failure-example.yaml
```

#### Warch your chaos experiments in Dashboard

Chaos Dashboard is currently only availble for TiDB clusters. Stay tuned for more supports or join us in making it happen.

**Note:** Make sure you have used the [option](#install-chaos-mesh) to deploy Chaos Mesh with Chaos Dashboard.

A typical way to access it is to use `kubectl port-forward`

```bash
kubectl port-forward -n chaos-testing svc/chaos-dashboard 8080:80
```

Then you can access [`http://localhost:8080`](http://localhost:8080) in browser.

## Roadmap

- [x] chaos-operator
- [ ] chaos-dashboard
- [ ] chaos-verify
- [ ] chaos-engine
- [ ] chaos-admin
- [ ] chaos-cloud

## License

Chaos Mesh is licensed under the Apache License, Version 2.0. See [LICENSE](/LICENSE) for the full license text.
