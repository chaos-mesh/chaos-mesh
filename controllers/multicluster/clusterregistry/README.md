# Multicluster controller registry

This package manages one in-process controller-runtime manager for each configured remote cluster. It is part of a larger multicluster flow involving:

- `remotecluster`, which reads a `RemoteCluster` resource, installs or upgrades Chaos Mesh in the remote cluster through Helm, and starts or stops the remote manager;
- `remotechaos`, which watches management-cluster chaos objects with a non-empty remote-cluster field and creates or deletes their remote copies;
- `clusterregistry`, which owns the remote Fx applications, managers, and clients; and
- `remotechaosmonitor`, which runs inside each remote manager and mirrors remote status/finalizers back to the management-cluster object.

The common local chaos pipeline filters out objects whose remote-cluster field is non-empty. The copied remote object has that field cleared, so the remote cluster's normal Chaos Mesh controllers execute it.

## Registry lifecycle

`RemoteClusterRegistry` exposes three operations:

- `Spawn(name, restConfig)` starts a remote manager;
- `Stop(ctx, name)` stops it and removes it from the registry;
- `WithClient(name, callback)` runs a callback with a live, uncached remote-cluster client.

Registrations are keyed by remote-cluster name. Starting, busy, or stopping entries reject client callbacks with `ErrBusy`; callbacks for other entries remain available. Starting and stopping names also reject `Spawn` with `ErrBusy`. `Spawn` returns `ErrAlreadyExist` rather than replacing a running manager, and `Stop` or `WithClient` returns `ErrNotExist` for an unknown name. A changed REST configuration is not applied to an existing entry; callers that need to replace it must stop the old manager before spawning another one.

The registry is provided from `controllers.Module`. The `remotecluster` reconciler obtains the REST configuration from the referenced kubeconfig Secret and controls the registry entry together with the remote Helm release and the `chaos-mesh/remotecluster-controllers` finalizer.

## Remote Fx application

`Spawn` constructs a separate Fx application for the remote cluster. It supplies:

- the remote `*rest.Config`;
- the remote cluster name as `name:"cluster-name"`;
- the main manager's client as `name:"manage-client"`;
- a controller-runtime scheme containing the registered Chaos Mesh objects; and
- an unnamed manager and client configured for the remote cluster.

The remote manager currently:

- disables metrics serving;
- disables leader election;
- does not register admission webhooks; and
- loads `remotechaosmonitor.Module` to watch remote chaos objects.

The Fx lifecycle starts the manager with a cancelable context and waits for it to stop during `app.Stop`.

## Client identities

Inside a remote-controller Fx module:

| Dependency | Cluster |
| --- | --- |
| unnamed `client.Client` | Remote cluster watched by that manager |
| `client.Client name:"manage-client"` | Main management cluster |
| `string name:"cluster-name"` | Registry key for the remote cluster |

Keep these identities explicit in Fx annotations. Mixing the two clients can create or delete resources in the wrong cluster.

From a management-cluster reconciler, use `WithClient` for a short operation against the remote cluster:

```go
err := registry.WithClient(clusterName, func(remoteClient client.Client) error {
    return remoteClient.Get(ctx, key, object)
})
```

`WithClient` serializes callbacks for one cluster and immediately rejects additional calls for that cluster with `ErrBusy`. The registry map lock is released before any callback or manager lifecycle work. Do not perform unbounded work, stop the same cluster, or recursively borrow it from a callback.

Remote chaos forwarding uses four workers per chaos kind and a ten-second reconciliation context. The registry copies the REST configuration and caps its HTTP timeout at ten seconds, preserving any stricter timeout. Forwarding reads bypass the manager cache; monitor controllers retain their existing cached client. API failures, unavailable registrations, and busy entries trigger rate-limited retries. Deletion is checked again after one second until a live remote read confirms absence before local finalizers are cleared.

One unavailable target cannot occupy all workers through waiting callbacks. Four distinct unavailable targets can still occupy the finite worker pool until their calls finish. Discovery may make multiple individually bounded requests, and arbitrary callback code must honor cancellation; the reconciliation deadline is not an unconditional wall-clock guarantee.

## Registering another remote reconciler

To add a controller that runs in every remote manager:

1. create a package with an Fx module that provides its reconciler and invokes its bootstrap;
2. accept the unnamed client for remote-cluster operations;
3. request the named management client or cluster name only when required;
4. add the module to the Fx application in `RemoteClusterRegistry.Spawn`, alongside `remotechaosmonitor.Module`; and
5. add tests that prove resources are read from and written to the intended cluster.

Use `controllers/multicluster/remotechaosmonitor/fx.go` as the active registration example. `remotepodreconciler.Module` exists in the tree but is not currently included in `Spawn`, so it is not an example of an active remote controller.

If the new controller requires webhooks, metrics, leader election, a no-cache reader, or another typed client, add those dependencies deliberately to the remote application; they are not inherited automatically from the main manager.

## Security and shutdown

The kubeconfig referenced by `RemoteCluster` must authorize the Helm operations performed by the management controller and every read/write/watch performed by remote reconcilers. Grant only the permissions the configured features require.

Stopping closes admission before waiting for the active callback, then cancels the remote manager and waits for it to exit. Each caller may stop waiting when its context expires, but the name stays reserved until actual shutdown completes. Concurrent callers join one shutdown operation. Successful late completion removes the entry; a shutdown error leaves it unavailable and returns the recorded error on subsequent `Stop` calls. It cannot be replaced automatically.

The manager starts in the Fx start hook, so failed construction does not launch it. Failed startup keeps the name reserved through lifecycle cleanup. Controller code must honor the reconcile context and avoid unbounded operations so shutdown can complete.

Focused validation:

```bash
go test ./controllers/multicluster/...
```
