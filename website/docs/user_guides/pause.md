---
id: pause_experiment 
title: Pause Chaos Experiment
sidebar_label: Pause Chaos Experiment
---

This document describes how to pause a running chaos experiment in Chaos Mesh.

Pause is a state suggests that a running chaos experiment has been temporarily hung up but not deleted.

Resume a paused experiment means running the chaos experiment again with the same parameters.

## Usage

Below is a sample configuration file of PodChaos:

```yaml
apiVersion: pingcap.com/v1alpha1
kind: PodChaos
metadata:
  name: pod-kill-example
  namespace: chaos-testing
spec:
  action: pod-kill
  mode: one
  selector:
    labelSelectors:
      "app.kubernetes.io/component": "chaos-daemon"
  duration: "10s"
  scheduler:
    cron: "@every 15s"
```

1. When the chaos is running, run the following command to get its status:

    ```shell
    $ kubectl get podchaos pod-kill-example --namespace chaos-testing --output yaml \
    && kubectl get pods --namespace chaos-testing
    ```

    The output is like this:

    ```
    ...
    spec:
      action: pod-kill
      containerName: ""
      duration: 10s
      mode: one
      nextRecover: "2020-04-15T03:18:14Z"
      nextStart: "2020-04-15T03:18:19Z"
      paused: false
      scheduler:
        cron: '@every 15s'
      selector:
        labelSelectors:
          app.kubernetes.io/component: chaos-daemon
        namespaces:
        - chaos-testing
      value: ""
    status:
      experiment:
        endTime: "2020-04-15T03:17:59Z"
        phase: Running
        podChaos:
          - action: pod-kill
          hostIP: 172.17.0.5
          message: delete pod
          name: chaos-daemon-mdwqr
          namespace: chaos-testing
          podIP: 10.244.2.3
        startTime: "2020-04-15T03:18:04Z"
      phase: ""
    NAME                                        READY   STATUS              RESTARTS   AGE
    chaos-controller-manager-7f67fbcfdc-ljlkn   1/1     Running             0          39s
    chaos-daemon-8cdv2                          1/1     Running             0          15s
    chaos-daemon-k7smn                          0/1     ContainerCreating   0          1s
    chaos-daemon-p9wxd                          1/1     Running             0          39s
    ```

2. Pause the running chaos:

    ```shell
    $  kubectl annotate podchaos pod-kill-example --namespace chaos-testing experiment.pingcap.com/pause=true
    podchaos.pingcap.com/pod-kill-example annotated
    $ kubectl get podchaos pod-kill-example --namespace chaos-testing --output yaml \
    && kubectl get pods --namespace chaos-testing
    ```

    The output is like this:

    ```
    ...
    metadata:
      annotations:
        experiment.pingcap.com/pause: "true"
    ...
    spec:
      action: pod-kill
      containerName: ""
      duration: 10s
      mode: one
      nextStart: "2020-04-15T03:18:34Z"
      paused: true
      scheduler:
        cron: '@every 15s'
      selector:
        labelSelectors:
          app.kubernetes.io/component: chaos-daemon
        namespaces:
        - chaos-testing
      value: ""
    status:
      experiment:
        endTime: "2020-04-15T03:18:24Z"
        phase: Paused
        podChaos:
        - action: pod-kill
          hostIP: 172.17.0.4
          message: delete pod
          name: chaos-daemon-p9wxd
          namespace: chaos-testing
          podIP: 10.244.3.3
        startTime: "2020-04-15T03:18:19Z"
      phase: ""
    NAME                                        READY   STATUS    RESTARTS   AGE
    chaos-controller-manager-7f67fbcfdc-ljlkn   1/1     Running   0          5m58s
    chaos-daemon-8cdv2                          1/1     Running   0          5m34s
    chaos-daemon-k7smn                          1/1     Running   0          5m20s
    chaos-daemon-sflc4                          1/1     Running   0          5m5s
    ```

3. Resume this chaos:

    ```shell
    $ kubectl annotate podchaos pod-kill-example --namespace chaos-testing experiment.pingcap.com/pause-
    podchaos.pingcap.com/pod-kill-example annotated
    ```

    The output is like this:

    ```
    ...
    spec:
      action: pod-kill
      containerName: ""
      duration: 10s
      mode: one
      nextRecover: "2020-04-15T03:23:56Z"
      nextStart: "2020-04-15T03:24:01Z"
      scheduler:
        cron: '@every 15s'
      selector:
        labelSelectors:
          app.kubernetes.io/component: chaos-daemon
        namespaces:
        - chaos-testing
      value: ""
    status:
      experiment:
        endTime: "2020-04-15T03:18:24Z"
        phase: Running
        podChaos:
        - action: pod-kill
          hostIP: 172.17.0.5
          message: delete pod
          name: chaos-daemon-k7smn
          namespace: chaos-testing
          podIP: 10.244.2.4
        startTime: "2020-04-15T03:23:46Z"
      phase: ""
    NAME                                        READY   STATUS    RESTARTS   AGE
    chaos-controller-manager-7f67fbcfdc-ljlkn   1/1     Running   0          6m29s
    chaos-daemon-2pcs9                          1/1     Running   0          9s
    chaos-daemon-8cdv2                          1/1     Running   0          6m5s
    chaos-daemon-sflc4                          1/1     Running   0          5m36s
    ```
