---
id: faqs
title: FAQs
sidebar_label: FAQs
---

## Question

### Q: If I do not have Kubernetes clusters deployed, can I use Chaos Mesh to create chaos experiments?

No, you can not use Chaos Mesh in this case. But still you can run chaos experiments using command line. Refer to [Command Line Usages of Chaos](https://github.com/pingcap/tipocket/blob/master/doc/command_line_chaos.md) for details.

### Q: When I deploy Chaos-mesh successfully, Then I create PodChaos experiments successfully, but create NetworkChaos/TimeChaos Experiment failed. The log is shown below, what's the matter？

I get chaos-controller-manager's log, as shown below.

```
2020-06-18T01:05:26.207Z	ERROR	controllers.TimeChaos	failed to apply chaos on all pods	{"reconciler": "timechaos", "error": "rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing dial tcp xx.xx.xx.xx:xxxxx: connect: connection refused\""}
```

You can try use the parameters: `hostNetwork`. Usage is as follows:

```
# vim helm/chaos-mesh/values.yaml, change hostNetwork from false to true
hostNetwork: true
```


## Debug

### Q: Experiment not working after chaos is applied

You can debug as described below:

Execute `kubectl describe` to check the specified chaos experiment resource.

- If there are `NextStart` and `NextRecover` fields under `spec`, then the chaos will be triggered after `NextStart` is executed.

- If there are no `NextStart` and `NextRecover`fields in `spec`, run the following command to get controller-manager's log and see whether there are errors in it.

    ```bash
    kubectl logs -n chaos-testing chaos-controller-manager-xxxxx (replace this with the name of the controller-manager) | grep "ERROR"
    ```

    For error message `no pod is selected`, run the following command to show the labels and check if the selector is desired.

    ```bash
    kubectl get pods -n yourNamespace --show-labels
    ```

If the above steps cannot solve the problem or you encounter other related errors in controller's log, [file an issue](https://github.com/pingcap/chaos-mesh/issues) or message us in #sig-chaos-mesh channel in the [TiDB Community](https://pingcap.com/tidbslack) slack workspace.

## IOChaos

### Q: Running chaosfs sidecar container failed, and log shows `pid file found, ensure docker is not running or delete /tmp/fuse/pid`

The chaosfs sidecar container is continuously restarting, and you might see the following logs at the current sidecar container:

```
2020-01-19T06:30:56.629Z	INFO	chaos-daemon	Init hookfs
2020-01-19T06:30:56.630Z	ERROR	chaos-daemon	failed to create pid file	{"error": "pid file found, ensure docker is not running or delete /tmp/fuse/pid"}
github.com/go-logr/zapr.(*zapLogger).Error
```

* **Cause**: Chaos Mesh uses Fuse to inject I/O failures. It fails if you specify an existing directory as the source path for chaos. This often happens when you try to reuse a persistent volume (PV) with the `Retain` reclaim policy to request a PersistentVolumeClaims (PVC) resource.
* **Solution**: In this case, use the following command to change the reclaim policy to `Delete`:

```bash
kubectl patch pv <your-pv-name> -p '{"spec":{"persistentVolumeReclaimPolicy":"Delete"}}'
```
