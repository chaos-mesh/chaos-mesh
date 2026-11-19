This image will be used to run `dlv` inside the chaos mesh pods. If Chaos Mesh
is configured properly, you can use it to debug the process on remote machine.

## Deploy

To deploy the `dlv` image, you need to compile the chaos-mesh with symbols:

```bash
make DEBUGGER=1 IMAGE_TAG=debug image
make -B IMAGE_TAG=debug image-chaos-dlv
```

The second command forces the Delve image to rebuild. Its `.dockerbuilt` stamp does not track `IMAGE_TAG`, so changing the tag alone can skip that image.

Load the built `chaos-mesh`, `chaos-daemon`, `chaos-dashboard`, and `chaos-dlv` images into the cluster nodes before installation. For a local minikube cluster:

```bash
minikube image load ghcr.io/chaos-mesh/chaos-mesh:debug
minikube image load ghcr.io/chaos-mesh/chaos-daemon:debug
minikube image load ghcr.io/chaos-mesh/chaos-dashboard:debug
minikube image load ghcr.io/chaos-mesh/chaos-dlv:debug
```

For a remote cluster, push the debug images to a registry reachable by every node and set `images.registry` when needed. Use the same image tag for the build and Helm installation. Choose a fresh tag for each debug build to avoid reusing cached images.

Install the chart with the debugging sidecars enabled:

```bash
helm install chaos-mesh ./helm/chaos-mesh -n chaos-mesh --create-namespace \
  --set chaosDlv.enable=true,images.tag=debug
```

Controller Manager, Chaos Daemon, and Dashboard Pods contain a `chaos-mesh-dlv` sidecar. DNS server Pods do not. List the Pod names and their containers:

```bash
kubectl get pods -n chaos-mesh -o jsonpath='{range .items[*]}{.metadata.name}{": "}{.spec.containers[*].name}{"\n"}{end}'
```

## Usage

### CLI

After deploying the chaos mesh with `dlv` support, you can use the `kubectl`
command to forward the remote port, and use a `dlv` command to connect to it.
For example:

```bash
kubectl port-forward -n chaos-mesh svc/chaos-mesh-controller-manager 2345:8000
dlv connect localhost:2345
```

Then you have entered the `dlv` debug environment and can debug the process.

### VSCode

A more fashion way would be configuring the Visual Studio Code tasks to attach
to specific pod automatically.

Firstly, use `kubectl port-forward` to export the port on local machine:

```bash
kubectl port-forward -n chaos-mesh svc/chaos-mesh-controller-manager 2345:8000
```

Then add a vscode attach configuration and save it in `.vscode/launch.json`:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Attach chaos-mesh",
            "type": "go",
            "request": "attach",
            "mode": "remote",
            "port": 2345,
            "host": "127.0.0.1",
            "apiVersion": 2,
            "showLog": true,
            "remotePath": "/mnt",
            "trace": "verbose"
        },
    ]
}
```

Click `Start Debugging` in "RUN AND DEBUG" section, and you will be able to
pause the process and see the backtrace, variables and set breakpoints.
