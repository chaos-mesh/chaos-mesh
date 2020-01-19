# FAQs

## IOChaos

### Q: chaosfs sidecar container run failed, and log shows `pid file found, ensure docker is not running or delete /tmp/fuse/pid`

```
2020-01-19T06:30:56.629Z	INFO	chaos-daemon	Init hookfs
2020-01-19T06:30:56.630Z	ERROR	chaos-daemon	failed to create pid file	{"error": "pid file found, ensure docker is not running or delete /tmp/fuse/pid"}
github.com/go-logr/zapr.(*zapLogger).Error
```

* Cause: Chaos-mesh uses fuse to hijack I/O operations, it will fail if there already exists the directory which chosen as the source path. This often happens when reuses a `Retain` reclaim policy PV when request a PVC resource.
* Solution: In this case, you could use below command to change the reclaim policy to `Delete`:

```bash
kubectl patch pv <your-pv-name> -p '{"spec":{"persistentVolumeReclaimPolicy":"Delete"}}'
```
