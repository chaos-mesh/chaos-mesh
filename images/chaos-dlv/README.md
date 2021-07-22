This image will be used to run `dlv` inside the chaos mesh pods. If Chaos Mesh
is configured properly, you can use it to debug the process on remote machine.

## Deploy

To deploy the `dlv` image, you need to compile the chaos-mesh with symbols:

```bash
make DEBUGGER=1 image
```

Then you need to install with the corresponding helm configuration to integrate
this container with the chaos mesh:

```bash
helm install chaos-mesh ./helm/chaos-mesh -n chaos-testing --set chaosDlv.enable=true
```

After deploying the image, you can find the `chaos-mesh-dlv` container under
every pods of chaos mesh. The command:

```bash
kubectl get pods -n chaos-testing -o jsonpath="{.items[*].spec.containers[*].name}"
```

Will print:

```bash
chaos-mesh chaos-mesh-dlv chaos-daemon chaos-mesh-dlv
```

## Usage

### CLI

After deploying the chaos mesh with `dlv` support, you can use the `kubectl`
command to forward the remote port, and use a `dlv` command to connect to it.
For example:

```bash
kubectl port-forward -n chaos-testing svc/chaos-mesh-controller-manager 2345:8000
dlv connect localhost:2345
```

Then you have entered the `dlv` debug environment and can debug the process.

### VSCode

A more fashion way would be configuring the Visual Studio Code tasks to attach
to specific pod automatically.

Firstly, use `kubectl port-forward` to export the port on local machine:

```bash
kubectl port-forward -n chaos-testing svc/chaos-mesh-controller-manager 2345:8000
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
