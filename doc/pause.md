# Pause Experiment

This document describe how to pause a running chaos in Chaos Mesh.

## Pause

Pause is a state that a running chaos is recovered temporally but not deleted.
Undoing pausing a chaos will run the chaos again with same parameter.

## How To

For instance we have a podchaos:
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
    namespaces:
      - chaos-testing
  duration: "10s"
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
  duration: 10s
  mode: one
  nextRecover: "2020-04-10T06:45:04Z"
  nextStart: "2020-04-10T06:45:09Z"
  paused: false
  scheduler:
    cron: '@every 15s'
  selector:
    namespaces:
    - chaos-testing
  value: ""
status:
  experiment:
    phase: Running
    podChaos:
    - action: pod-kill
      hostIP: 172.17.0.2
      message: delete pod
      name: chaos-daemon-g4s25
      namespace: chaos-testing
      podIP: ""
    startTime: "2020-04-10T06:44:54Z"
  phase: ""
NAME                                        READY   STATUS              RESTARTS   AGE
chaos-controller-manager-7f67fbcfdc-n2dht   1/1     Running             0          47s
chaos-daemon-2tfcw                          0/1     ContainerCreating   0          2s
chaos-daemon-5qq74                          1/1     Running             0          47s
chaos-daemon-bk4jd                          1/1     Running             0          47s
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
  duration: 10s
  mode: one
  nextStart: "2020-04-10T06:45:09Z"
  paused: true
  scheduler:
    cron: '@every 15s'
  selector:
    namespaces:
    - chaos-testing
  value: ""
status:
  experiment:
    endTime: "2020-04-10T06:45:03Z"
    phase: Paused
    podChaos:
    - action: pod-kill
      hostIP: 172.17.0.2
      message: delete pod
      name: chaos-daemon-g4s25
      namespace: chaos-testing
      podIP: ""
    startTime: "2020-04-10T06:44:54Z"
  phase: ""
NAME                                        READY   STATUS    RESTARTS   AGE
chaos-controller-manager-7f67fbcfdc-n2dht   1/1     Running   0          60s
chaos-daemon-2tfcw                          1/1     Running   0          15s
chaos-daemon-5qq74                          1/1     Running   0          60s
chaos-daemon-bk4jd                          1/1     Running   0          60s
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
  duration: 10s
  mode: one
  nextRecover: "2020-04-10T06:45:27Z"
  nextStart: "2020-04-10T06:45:32Z"
  paused: false
  scheduler:
    cron: '@every 15s'
  selector:
    namespaces:
    - chaos-testing
  value: ""
status:
  experiment:
    endTime: "2020-04-10T06:45:03Z"
    phase: Running
    podChaos:
    - action: pod-kill
      hostIP: 172.17.0.2
      message: delete pod
      name: chaos-controller-manager-7f67fbcfdc-n2dht
      namespace: chaos-testing
      podIP: 10.244.2.5
    startTime: "2020-04-10T06:45:18Z"
  phase: ""
NAME                                        READY   STATUS    RESTARTS   AGE
chaos-controller-manager-7f67fbcfdc-mdh7l   1/1     Running   0          7s
chaos-daemon-2tfcw                          1/1     Running   0          31s
chaos-daemon-5qq74                          1/1     Running   0          76s
chaos-daemon-bk4jd                          1/1     Running   0          76s
```