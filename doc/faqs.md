# FAQs

## IOChaos

### Q: chaosfs sidecar container run failed, and log shows `pid file found, ensure docker is not running or delete /tmp/fuse/pid`

```
2020-01-19T06:30:56.629Z	INFO	chaos-daemon	Init hookfs
2020-01-19T06:30:56.630Z	ERROR	chaos-daemon	failed to create pid file	{"error": "pid file found, ensure docker is not running or delete /tmp/fuse/pid"}
github.com/go-logr/zapr.(*zapLogger).Error
```

* Cause: Chaos-mesh uses fuse to hijack I/O operations. It fails if you specify an existing directory as the source path for chaos. This often happens when you try to reuse a persistent volume (PV) with the `Retain` reclaim policy to request a PVC resource.
* Solution: In this case, use the following command to change the reclaim policy to `Delete`:

```bash
kubectl patch pv <your-pv-name> -p '{"spec":{"persistentVolumeReclaimPolicy":"Delete"}}'
```
