# FAQs

## Debug

### Q: Experiment not working after apply chaos

You can following the steps to debug:
- First using `kubectl describe` the chaos experiment resource. If see `NextStart`  and `NextRecover`  field in spec, then after `NextStart` the chaos will trigger.
- Then use `kubectl logs -n chaos-testing chaos-controller-manager-xxxxx(replace this will controller-manager's pod name) | grep "ERROR"`  to get controller-manager's log and see whether there are errors in it. If have the error message like `no pod is selected`, then you can can use `kubectl get pods -n yourNamespace(using your namespace) --show-labels` to show the labels and check if it is meet selector. If there are some other related errors in controller's log, please file an issue.
- If above can not solve the problem, please contract us by file an issue or direct message in slack channel

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
