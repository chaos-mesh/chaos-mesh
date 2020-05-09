# FAQs

## Question

### Q: If I do not have k8s, can I using Chaos Mesh to create chaos experiments

No, you cannot use Chaos Mesh in this case, but you can still do chaos experiments using command line following [Command Line Usages of Chaos](https://github.com/pingcap/tipocket/blob/master/doc/command_line_chaos.md)

## Debug

### Q: Experiment not working after chaos is applied

You can debug as described below:

Use `kubectl describe` to show the specified chaos experiment resource.

- If there are `NextStart` and `NextRecover` fields under `spec`, then the chaos will be triggered after `NextStart` is executed.

- If there are no `NextStart` and `NextRecover`fields in `spec`, then use `kubectl logs -n chaos-testing chaos-controller-manager-xxxxx (replace this with the name of the controller-manager) | grep "ERROR"` to get controller-manager's log and see whether there are errors in it. For error message `no pod is selected`, use `kubectl get pods -n yourNamespace --show-labels` to show the labels and check if the selector is desired. For other related errors in controller's log, please file an issue.

If the above steps cannot solve the problem, please contact us by filing an issue or message us in the slack channel.

## IOChaos

### Q: chaosfs sidecar container run failed, and log shows `pid file found, ensure docker is not running or delete /tmp/fuse/pid`

The chaosfs sidecar container is continuously restarting, and you may see following logs at the current sidecar container:

```
2020-01-19T06:30:56.629Z	INFO	chaos-daemon	Init hookfs
2020-01-19T06:30:56.630Z	ERROR	chaos-daemon	failed to create pid file	{"error": "pid file found, ensure docker is not running or delete /tmp/fuse/pid"}
github.com/go-logr/zapr.(*zapLogger).Error
```

* **Cause**: chaos-mesh uses fuse to hijack I/O operations. It fails if you specify an existing directory as the source path for chaos. This often happens when you try to reuse a persistent volume (PV) with the `Retain` reclaim policy to request a PersistentVolumeClaims (PVC) resource.
* **Solution**: In this case, use the following command to change the reclaim policy to `Delete`:

```bash
kubectl patch pv <your-pv-name> -p '{"spec":{"persistentVolumeReclaimPolicy":"Delete"}}'
```
