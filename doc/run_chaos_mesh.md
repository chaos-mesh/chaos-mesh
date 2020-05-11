# Run Chaos Mesh

Now you have deployed Chaos Mesh in your environment, it's time 
to use it for your chaos experiments. This document walks you through the process of running chaos experiments. It also describes the regular operations on chaos experiments.  

## Step 1: Deploy the target cluster

The first step is always to have the target cluster to test deployed. For illustration purposes, TiDB is used as a sample cluster.

You can follow the instructions in the following two documents to deploy a TiDB cluster:

* [Deploy using kind](https://pingcap.com/docs/tidb-in-kubernetes/stable/deploy-tidb-from-kubernetes-kind/)
* [Deploy using minikube](https://pingcap.com/docs/tidb-in-kubernetes/stable/deploy-tidb-from-kubernetes-minikube/)

## Step 2: Define the experiment config file

The chaos experiment configuration is defined in a `.yaml` file. You need to create your own experiment config file, based on the available fields in the sample below:

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
    labelSelectors:
      "app.kubernetes.io/component": "tikv" # the label of the pod for chaos injection
  scheduler: # scheduler rules for the running time of the chaos experiments about pods.
    cron: "@every 5m"
```

## Step 3: Apply a chaos experiment

Running the following commands to apply the experiment config defined in the `.yaml` file:

```bash
kubectl apply -f pod-failure-example.yaml
kubectl get podchaos --namespace=chaos-testing
```

With this step, you now run your chaos experiment successfully. By [running a benchmark against the cluster](https://pingcap.com/docs/stable/benchmark/how-to-run-sysbench/), you can notice the QPS performance affected by the chaos experiment:

![tikv-pod-failure](../static/tikv-pod-failure.png)

## Regular operations on chaos experiments

In this section, you will learn about some follow-up operations on a chaos experiment after it is applied.

### Update a chaos experiment

```bash
vim pod-failure-example.yaml # modify pod-failure-example.yaml to what you want
kubectl apply -f pod-failure-example.yaml
```

### Delete a chaos experiment

```bash
kubectl delete -f pod-failure-example.yaml
```
