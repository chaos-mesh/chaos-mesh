# Multicluster Technical Documents

This documents not only describe the behavior of `clusterregistry`, it
also describes the technical framework which can help the chaos mesh developers
to develop the multicluster application.

A `clusterregistry` will manage all controllers watching a remote cluster. The
construction of these controllers (managers) will be managed by `fx`. The main
process of constructing a controller manage is nearly the same with the main
one. The only difference is that we'll need to provide a new `RestConfig`, and
`Populate` the client to allow others to use its client.

The `RemoteClusterRegistry` provides three methods: `Spawn`, `Stop` and
`WithClient`. `Spawn` allows you to setup a new controller-manager watching
resources inside the remote cluster, and `Stop` allows you to stop a running
controller-manager. `WithClient` enables the developer to get a client to
operate in the remote cluster.

For more details about these three functions, please read the documents of them.

## Bootstrap Process

You'll need a `*rest.Config` to start a controller manager inside remote
cluster. This config is used to setup client, watch changes... This client will
also be used by any other called `WithClient`, so make sure it has enough
priviledges.

With this `*rest.Config`, the `Spawn` method will construct a new fx App, which
is nearly the same with the main one. The difference between this fx App and the
main one is that it only adds reconciler which is needed by multicluster, and it
doesn't start webhook, doesn't listen on the metrics.

Except `ctrl.Option` and `*rest.Config`, all other arguments used by manager are
provided by the same one in the `provider`. Only the default client is passed
inside. If you need more different client, it won't be too complicated to add
them to this construction process.

It also passes a cancelable context to the `run` function. This context is used
to stop the controller.

## Stop Process

The stop / cancel is managed directly by the fx lifecycle. We added a stop hook
to cancel the context used by the controller manager when we are stopping the fx
app.

## Register Reconciler

If you need to register a reconciler for a remote cluster, add a new `fx.Invoke`
to the construction of cluster (in the `Spawn` method), register the new
reconciler to the manager inside that function.

See `/controllers/multicluster/remotepodreconciler/fx.go` for an example.

## Use the remote cluster client in manage cluster reconciler

To manage the resources in the remote cluster, you may need to get a client of
remote cluster. With a `ClusterRegistry`, it would be really easy!

```go
err := r.registry.WithClient(obj.Name, func(c client.Client) error {
    return c.Create(ctx, &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "hello-world",
            Namespace: "default",
        },
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {
                    Image:           "docker/whalesay",
                    ImagePullPolicy: corev1.PullIfNotPresent,
                    Name:            "hello-world",
                    Command:         []string{"cowsay", "Hello World"},
                },
            },
        },
    })
})
if err != nil {
    if !k8sError.IsAlreadyExists(err) {
        r.Log.Error(err, "fail to create pod")
    }

}
```

## Use the manage cluster client in remote cluster reconciler

`ClusterRegistry` provides a `client.Client` annotated with
`name:"manage-client"` to allow remote cluster reconciler operates on the manage
cluster. See `/controllers/multicluster/remotepodreconciler/fx.go` for an
example.
