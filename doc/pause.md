# Pause Experiment

This document describes how to pause a running chaos in Chaos Mesh.

## Pause

Pause is a state a running chaos has been recovered temporally but not deleted.
Undoing pausing a chaos will run the chaos again with same parameter.

## How To

For instance, we have a podchaos:
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
  scheduler:
    cron: "@every 15s"
```

While the chaos is running we can get its status like below:
```shell script
$ kubectl get podchaos pod-kill-example --namespace chaos-testing --output yaml \
&& kubectl get pods --namespace chaos-testing
...
spec:
  action: pod-kill
  containerName: ""
  mode: one
  nextStart: "2020-04-15T03:11:00Z"
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
    endTime: "2020-04-15T03:10:45Z"
    phase: Finished
    podChaos:
    - action: pod-kill
      hostIP: 172.17.0.5
      message: delete pod
      name: chaos-daemon-j8n7h
      namespace: chaos-testing
      podIP: 10.244.2.3
    startTime: "2020-04-15T03:10:45Z"
  phase: ""
NAME                                        READY   STATUS    RESTARTS   AGE
chaos-controller-manager-7f67fbcfdc-n4rps   1/1     Running   0          25s
chaos-daemon-6ssph                          1/1     Running   0          7s
chaos-daemon-8rsvv                          1/1     Running   0          25s
chaos-daemon-qq6cp                          1/1     Running   0          25s
```

Pause the running chaos:
```shell script
$ kubectl patch podchaos pod-kill-example --namespace chaos-testing --type merge --patch 'spec:
  paused: true'
podchaos.pingcap.com/pod-kill-example patched
$ kubectl get podchaos pod-kill-example --namespace chaos-testing --output yaml \
&& kubectl get pods --namespace chaos-testing
...
spec:
  action: pod-kill
  containerName: ""
  mode: one
  nextStart: "2020-04-15T03:11:15Z"
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
    endTime: "2020-04-15T03:11:00Z"
    phase: Paused
    podChaos:
    - action: pod-kill
      hostIP: 172.17.0.3
      message: delete pod
      name: chaos-daemon-8rsvv
      namespace: chaos-testing
      podIP: 10.244.3.3
    startTime: "2020-04-15T03:11:00Z"
  phase: ""
NAME                                        READY   STATUS    RESTARTS   AGE
chaos-controller-manager-7f67fbcfdc-n4rps   1/1     Running   0          54s
chaos-daemon-6ssph                          1/1     Running   0          36s
chaos-daemon-qq6cp                          1/1     Running   0          54s
chaos-daemon-rjl94                          1/1     Running   0          21s
```

Resume this chaos:
```shell script
$ kubectl patch podchaos pod-kill-example --namespace chaos-testing --type merge --patch 'spec:
 paused: false'
podchaos.pingcap.com/pod-kill-example patched 
$ kubectl get podchaos pod-kill-example --namespace chaos-testing --output yaml \
&& kubectl get pods --namespace chaos-testing
...
spec:
  action: pod-kill
  containerName: ""
  mode: one
  nextStart: "2020-04-15T03:11:44Z"
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
    endTime: "2020-04-15T03:11:29Z"
    phase: Finished
    podChaos:
    - action: pod-kill
      hostIP: 172.17.0.3
      message: delete pod
      name: chaos-daemon-rjl94
      namespace: chaos-testing
      podIP: 10.244.3.5
    startTime: "2020-04-15T03:11:29Z"
  phase: ""
NAME                                        READY   STATUS    RESTARTS   AGE
chaos-controller-manager-7f67fbcfdc-n4rps   1/1     Running   0          67s
chaos-daemon-6ssph                          1/1     Running   0          49s
chaos-daemon-6t9tb                          1/1     Running   0          5s
chaos-daemon-qq6cp                          1/1     Running   0          67s
```