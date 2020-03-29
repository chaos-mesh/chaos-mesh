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
status:
  experiment:
    endTime: "2020-03-29T08:26:46Z"
    phase: Running
    podChaos:
    - action: pod-kill
      hostIP: 172.17.0.2
      message: delete pod
      name: chaos-controller-manager-7f67fbcfdc-lz9jh
      namespace: chaos-testing
      podIP: ""
    startTime: "2020-03-29T08:26:52Z"
  paused: false
  phase: ""
NAME                                        READY   STATUS              RESTARTS   AGE
chaos-controller-manager-7f67fbcfdc-nm6jl   1/1     Running             0          8s
chaos-daemon-fbpsz                          1/1     Running             0          5m56s
chaos-daemon-z7p5f                          1/1     Running             0          5m26s
chaos-daemon-zhz2k                          0/1     ContainerCreating   0          6s
```

Pause the running chaos:
```shell script
$ kubectl patch podchaos pod-kill-example --namespace chaos-testing --type merge --patch 'status:
  paused: true'
podchaos.pingcap.com/pod-kill-example patched
$ kubectl get podchaos pod-kill-example --namespace chaos-testing --output yaml \
&& kubectl get pods --namespace chaos-testing
...
status:
  experiment:
    endTime: "2020-03-29T08:28:26Z"
    phase: Paused
    podChaos:
    - action: pod-kill
      hostIP: 172.17.0.3
      message: delete pod
      name: chaos-daemon-scnw2
      namespace: chaos-testing
      podIP: 10.244.3.29
    startTime: "2020-03-29T08:28:22Z"
  paused: true
  phase: ""
NAME                                        READY   STATUS    RESTARTS   AGE
chaos-controller-manager-7f67fbcfdc-bx5vs   1/1     Running   0          26s
chaos-daemon-2nw8x                          1/1     Running   0          11s
chaos-daemon-4dzh4                          1/1     Running   0          71s
chaos-daemon-dlmfb                          1/1     Running   0          41s
```

Resume this chaos:
```shell script
$ kubectl patch podchaos pod-kill-example --namespace chaos-testing --type merge --patch 'status:
 paused: false'
podchaos.pingcap.com/pod-kill-example patched 
$ kubectl get podchaos pod-kill-example --namespace chaos-testing --output yaml \
&& kubectl get pods --namespace chaos-testing
...
status:
  experiment:
    endTime: "2020-03-29T08:28:26Z"
    phase: Running
    podChaos:
    - action: pod-kill
      hostIP: 172.17.0.3
      message: delete pod
      name: chaos-daemon-2nw8x
      namespace: chaos-testing
      podIP: 10.244.3.31
    startTime: "2020-03-29T08:54:22Z"
  paused: false
  phase: ""
NAME                                        READY   STATUS              RESTARTS   AGE
chaos-controller-manager-7f67fbcfdc-bx5vs   1/1     Running             0          26m
chaos-daemon-4dzh4                          1/1     Running             0          27m
chaos-daemon-dlmfb                          1/1     Running             0          26m
chaos-daemon-nmtdb                          0/1     ContainerCreating   0          2s
```