# Multicluster controller registry

This package manages one in-process controller-runtime manager for each configured remote cluster. It is part of a larger multicluster flow involving:

- `remotecluster`, which reads a `RemoteCluster` resource, installs or upgrades Chaos Mesh in the remote cluster through Helm, and starts or stops the remote manager;
- `remotechaos`, which watches management-cluster chaos objects with a non-empty remote-cluster field and creates or deletes their remote copies;
- `clusterregistry`, which owns the remote Fx applications, managers, and clients; and
- `remotechaosmonitor`, which runs inside each remote manager and mirrors remote status/finalizers back to the management-cluster object.

The common local chaos pipeline filters out objects whose remote-cluster field is non-empty. The copied remote object has that field cleared, so the remote cluster's normal Chaos Mesh controllers execute it.

## Registry lifecycle

`RemoteClusterRegistry` exposes these operations:

- `Ensure(ctx, name, kubeconfig)` starts a manager or replaces it when the decoded kubeconfig changes;
- `Spawn(name, restConfig)` starts a manager only when its name is unregistered;
- `Stop(ctx, name)` stops it and removes it from the registry;
- `WithClient(name, callback)` runs a callback with a live, uncached remote-cluster client.

Registrations are keyed by remote-cluster name. `Ensure` retains an independent kubeconfig snapshot and leaves an equal configuration running only when its manager is ready. It validates changed configuration before stopping the old manager and starting its replacement. Invalid configuration leaves the old manager running. Failed startup permits retry after successful cleanup; a shutdown error keeps the name unavailable.

Starting or stopping entries reject `Ensure`, `Spawn`, and client callbacks with `ErrBusy`. A completed shutdown error is returned by later `Ensure` and `Stop` calls. `Spawn` returns `ErrAlreadyExist` for a running manager. `Stop` or `WithClient` returns `ErrNotExist` for an unknown name. Same-cluster callbacks run one at a time, while other clusters remain available.

The registry is provided from `controllers.Module`. The `remotecluster` reconciler ensures its cleanup finalizer before external operations and calls `Ensure` for every live resource, rebuilding missing managers after a management-controller restart. Helm install/upgrade runs only for an unobserved `RemoteCluster` generation. Referenced Secret changes enqueue reconciliation and refresh the connection; changing a Secret endpoint alone does not migrate a Helm installation. Deletion stops the manager before reading credentials for Helm cleanup and removes only the owned finalizer after cleanup succeeds.

`RemoteCluster` reports `Ready=True` after setup for the current generation and manager registration succeed. Setup failures set `Ready=False` with a stable reason while preserving the last successful installation status and observed generation. Deletion sets `Ready=False` before stopping the manager, including when cleanup is blocked. `Ready` records the latest reconciliation result; it does not track continuous API connectivity, cache synchronization, or remote workload health.

## Remote Fx application

The shared startup path constructs a separate Fx application for the remote cluster. It supplies:

- the remote `*rest.Config`;
- the remote cluster name as `name:"cluster-name"`;
- the main manager's client as `name:"manage-client"`;
- a controller-runtime scheme containing the registered Chaos Mesh objects; and
- an unnamed manager and client configured for the remote cluster.

The remote manager currently:

- disables metrics serving;
- disables leader election;
- permits reuse of stable controller names after a manager stops;
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

Remote chaos forwarding uses four workers per chaos kind and a ten-second reconciliation context. The forwarding client uses its own HTTP client and discovery mapper with a ten-second timeout, preserving any stricter configured timeout. Forwarding reads bypass the manager cache. The manager and its monitor controllers retain the original REST timeout and cached client. API failures, unavailable registrations, and busy entries trigger rate-limited retries. Deletion is checked again after one second. Local finalizers remain until a live remote read confirms absence or the remote experiment is no longer owned by forwarding. Detachment rechecks the management object identity, target cluster, and deletion state.

One unavailable target cannot occupy all workers through waiting callbacks. Four distinct unavailable targets can still occupy the finite worker pool until their calls finish. Discovery may make multiple individually bounded requests, and arbitrary callback code must honor cancellation; the reconciliation deadline is not an unconditional wall-clock guarantee.

## Registering another remote reconciler

To add a controller that runs in every remote manager:

1. create a package with an Fx module that provides its reconciler and invokes its bootstrap;
2. accept the unnamed client for remote-cluster operations;
3. request the named management client or cluster name only when required;
4. add the module to the Fx application in the registry startup path, alongside `remotechaosmonitor.Module`; and
5. add tests that prove resources are read from and written to the intended cluster.

Use `controllers/multicluster/remotechaosmonitor/fx.go` as the active registration example. `remotepodreconciler.Module` exists in the tree but is not currently included in the remote application, so it is not an example of an active remote controller.

If the new controller requires webhooks, metrics, leader election, a no-cache reader, or another typed client, add those dependencies deliberately to the remote application; they are not inherited automatically from the main manager.

## Security and shutdown

The kubeconfig referenced by `RemoteCluster` must authorize the Helm operations performed by the management controller and every read/write/watch performed by remote reconcilers. Grant only the permissions the configured features require.

Stopping closes admission before waiting for the active callback, then cancels the remote manager and waits for it to exit. Each caller may stop waiting when its context expires, but the name stays reserved until actual shutdown completes. Concurrent callers join one shutdown operation. Successful late completion removes the entry; a shutdown error leaves it unavailable and returns the recorded error on subsequent `Stop` calls. It cannot be replaced automatically.

The manager starts in the Fx start hook, so failed construction does not launch it. Failed startup keeps the name reserved through lifecycle cleanup. Controller code must honor the reconcile context and avoid unbounded operations so shutdown can complete.

Focused validation:

```bash
go test ./controllers/multicluster/...
```
