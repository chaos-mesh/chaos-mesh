---
id: run_chaos_experiment
title: Run Chaos Experiment
sidebar_label: Run Chaos Experiment
---

Now that you have deployed Chaos Mesh in your environment, it's time to use it for your chaos experiments. This document walks you through the process of running chaos experiments. It also describes the regular operations on chaos experiments.

## Step 1: Deploy the target cluster

The first step is always to deploy a testing cluster. For illustration purposes, TiDB is used as a sample cluster.

You can follow the instructions in the following two documents to deploy a TiDB cluster:

* [Deploy using kind](https://pingcap.com/docs/tidb-in-kubernetes/stable/deploy-tidb-from-kubernetes-kind/)
* [Deploy using Minikube](https://pingcap.com/docs/tidb-in-kubernetes/stable/deploy-tidb-from-kubernetes-minikube/)

## Step 2: Define the experiment configuration file

The chaos experiment configuration is defined in a YAML file. You need to create your own experiment configuration file based on the available fields in the sample below:

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

Run the following commands to apply the experiment:

```bash
kubectl apply -f pod-failure-example.yaml
kubectl get podchaos --namespace=chaos-testing
```

By [running a benchmark against the cluster](https://pingcap.com/docs/stable/benchmark/how-to-run-sysbench/), you can check the QPS performance affected by the chaos experiment:

![tikv-pod-failure](/img/tikv-pod-failure.png)

## Regular operations on chaos experiments

In this section, you will learn some follow-up operations when the chaos experiment is running.

### Update a chaos experiment

```bash
vim pod-failure-example.yaml # modify pod-failure-example.yaml to what you want
kubectl apply -f pod-failure-example.yaml
```

### Delete a chaos experiment

```bash
kubectl delete -f pod-failure-example.yaml
```

### Watch your chaos experiments in Chaos Dashboard

Chaos Dashboard is currently only available for TiDB clusters. Stay tuned for more supports or join us in making it happen.

> **Note:**
>
> If Chaos Dashboard was not installed, upgrade Chaos Mesh by executing `helm upgrade chaos-mesh helm/chaos-mesh --namespace=chaos-testing --set dashboard.create=true`.

A typical way to access it is to use `kubectl port-forward`:

```bash
kubectl port-forward -n chaos-testing svc/chaos-dashboard 8080:80
```

Then you can access [`http://localhost:8080`](http://localhost:8080) in the browser.
