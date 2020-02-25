# Run Chaos Mesh

Now that you have deployed Chaos Mesh in your environment, in this document, 
you will learn about how to use it for your chaos experiments.

## Step 1: Deploy target cluster

After Chaos Mesh is deployed, you can deploy the target cluster to be tested. For illustration purposes,  TiDB is used as a sample cluster.

You can follow the instructions in the following two documents to deploy a TiDB cluster:

* [Deploy using kind](https://pingcap.com/docs/stable/tidb-in-kubernetes/get-started/deploy-tidb-from-kubernetes-kind/)
* [Deploy using minikube](https://pingcap.com/docs/stable/tidb-in-kubernetes/get-started/deploy-tidb-from-kubernetes-minikube/)

## Step 2: Define the experiment config file

The chaos experiment configuration is defined in a ·`.yaml` file. In the following sample file, 
`pod-kill-example.yaml` defines a chaos experiment to kill one random TiKV pod every 60 seconds:

```yaml
apiVersion: pingcap.com/v1alpha1
kind: PodChaos
metadata:
  name: pod-failure-example
  namespace: chaos-testing
spec:
  action: pod-failure # the specific chaos action to inject; supported actions: pod-kill/pod-failure
  mode: one # the mode to run chaos action; supported modes are one/all/fixed/fixed-percent/random-max-percent
  duration: "60s" # duration for the injected chaos experiment
  selector: # pods where to inject chaos actions
    namespaces:
      - tidb-cluster-demo  # the namespace of the system under test (SUT) you've deployed
    labelSelectors:
      "app.kubernetes.io/component": "tikv" # the label of the pod for chaos injection
  scheduler: # scheduler rules for the running time of the chaos experiments about pods.
    cron: "@every 5m"
```

## Step 3: Create a chaos experiment

```bash
kubectl apply -f pod-failure-example.yaml
kubectl get podchaos --namespace=chaos-testing
```

You can see the QPS performance (by [running a benchmark against the cluster](https://pingcap.com/docs/stable/benchmark/how-to-run-sysbench/) affected by the chaos experiment from TiDB Grafana dashboard:

![tikv-pod-failure](../static/tikv-pod-failure.png)

## Step 4: Update a chaos experiment

```bash
vim pod-failure-example.yaml # modify pod-failure-example.yaml to what you want
kubectl apply -f pod-failure-example.yaml
```

## Step 5: Delete a chaos experiment

```bash
kubectl delete -f pod-failure-example.yaml
```

## Step 6: Watch your chaos experiments in Dashboard

Chaos Dashboard is currently only available for TiDB clusters. Stay tuned for more supports or join us in making it happen.

> **Note:**
>
> If Chaos Dashboard was not installed in your earlier deployment, you need to install it by upgrading Chaos Mesh:
>
> ```helm upgrade chaos-mesh helm/chaos-mesh --namespace=chaos-testing --set dashboard.create=true```

A typical way to access it is to use `kubectl port-forward`

```bash
kubectl port-forward -n chaos-testing svc/chaos-dashboard 8080:80
```

Then you can access [`http://localhost:8080`](http://localhost:8080) in browser.
