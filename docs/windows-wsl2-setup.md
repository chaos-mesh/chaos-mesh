# Setting up Chaos Mesh on Windows (WSL2 + Docker Desktop + Kind)

Scope
-----
This guide helps Windows users set up a local development environment for Chaos Mesh using WSL2, Docker Desktop, kind, kubectl, and Helm. It focuses on common Windows-specific issues and concise troubleshooting steps for beginners.

Prerequisites
-------------
- Windows 10/11 with WSL2 enabled and a Linux distro installed (Ubuntu recommended).
- Docker Desktop for Windows installed and configured to use WSL2 backend.
- `kubectl` installed in WSL2 and configured in PATH.
- `kind` (Kubernetes in Docker) installed inside WSL2.
- `helm` installed in WSL2.

High-level steps
----------------
1. Ensure WSL2 and Docker Desktop integration are working.
2. Create a kind cluster using the Docker Desktop runtime.
3. Install Chaos Mesh via Helm (or manifests) into the kind cluster.
4. Verify connectivity and troubleshoot common failures.

1. Common setup issues on Windows + WSL
--------------------------------------
- WSL distro not set to WSL2: run `wsl -l -v` in PowerShell to confirm the version for your distro.
  - Fix: `wsl --set-version <distro> 2`
- Docker Desktop not using WSL2 backend: open Docker Desktop Settings → General and enable "Use the WSL 2 based engine".
- File permission and path issues when mixing Windows and WSL paths. Operate from the Linux filesystem (e.g., `/home/<user>/projects`) rather than `C:\` mounted paths for best results.

2. Docker Desktop integration issues
------------------------------------
- Ensure Docker is accessible from WSL2: inside WSL run:

```sh
docker version
```

- If Docker commands fail inside WSL, enable integration for your distro in Docker Desktop Settings → Resources → WSL Integration.
- Restart Docker Desktop and WSL after changing integration settings:

PowerShell:
```powershell
wsl --shutdown
Restart-Service com.docker.service
```

3. Kind cluster creation problems
---------------------------------
- Use a minimal kind config that relies on Docker Desktop. Example `kind-config.yaml`:

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 30080
        hostPort: 30080
        protocol: TCP
```

- Create the cluster in WSL2:
```sh
kind create cluster --config kind-config.yaml --name chaos-mesh
```

- Common failures:
  - "pod sandbox changed" or CNI failures: ensure Docker Desktop Kubernetes is disabled (kind manages cluster), and that no port conflicts exist.
  - Slow image pulls: set `--image` to a local registry or pre-pull images.

4. kubectl connection issues
----------------------------
- Verify `kubectl` is using the kind cluster:

```sh
kubectl cluster-info
kubectl get nodes
```

- If kubectl cannot connect, check the kubeconfig in `~/.kube/config` and ensure `KUBECONFIG` is not overridden by a Windows path.
- To explicitly use kind's kubeconfig:

```sh
export KUBECONFIG="$(kind get kubeconfig-path --name=chaos-mesh)"
kubectl get nodes
```

5. RBAC token generation troubleshooting
----------------------------------------
- When following examples that require a service account token, create a token correctly in Kubernetes 1.24+ (TokenRequest API) or use a secret for older clusters.

Example: create a debug service account and retrieve a token:

```sh
kubectl create serviceaccount chaos-debug -n kube-system
kubectl create clusterrolebinding chaos-debug-binding --clusterrole=cluster-admin --serviceaccount=kube-system:chaos-debug
kubectl get secret -n kube-system $(kubectl get sa chaos-debug -n kube-system -o jsonpath="{.secrets[0].name}") -o go-template='{{.data.token | base64decode}}'
```

- If the secret is empty, use the TokenRequest API (recommended in newer Kubernetes releases):

```sh
kubectl -n kube-system create token chaos-debug
```

6. Port-forward troubleshooting
-------------------------------
- Use `kubectl port-forward` from inside WSL to avoid Windows/WSL networking quirks.

Example:
```sh
kubectl -n chaos-testing port-forward svc/chaos-dashboard 2333:2333
```

- If the port-forward fails with `error forwarding port`, ensure the requested host port is free on Windows and not blocked by firewall. Use an unprivileged port (>1024) if possible.

7. "No pods found" issue explanation
-----------------------------------
- Symptoms: `kubectl get pods -n chaos-mesh` returns no pods after install.
- Common causes and fixes:
  - Helm install targeted wrong namespace: verify `helm install -n chaos-mesh` and `kubectl get ns`.
  - Pod image pull failures: `kubectl describe pod <name>` and `kubectl get events -n chaos-mesh` reveal image pull errors.
  - CrashLoopBackOff due to missing permissions: check `kubectl logs` and `kubectl describe` for RBAC or admission webhook errors.

8. How to verify Chaos Mesh is working correctly
------------------------------------------------
- Check CRDs and controller status:

```sh
kubectl get crds | grep chaos
kubectl -n chaos-mesh get pods
kubectl -n chaos-mesh get deployments
```

- Deploy a simple PodChaos example and observe behavior:
```sh
kubectl apply -f examples/pod-kill-example.yaml
kubectl -n default get pod -w
kubectl describe podchaos podchaos-sample -n default
```

- Verify dashboard access (if installed) via port-forward and confirm UI shows the installed controllers.

9. Useful kubectl debug commands
-------------------------------
- Describe pods and view events:
```sh
kubectl describe pod <pod-name> -n <ns>
kubectl get events -n <ns> --sort-by='.lastTimestamp'
```
- Fetch logs for recent crashes:
```sh
kubectl logs <pod> -c <container> -n <ns> --previous
kubectl logs <pod> -c <container> -n <ns> --since=10m
```
- Exec into a running pod for live debugging:
```sh
kubectl exec -it <pod> -n <ns> -- /bin/sh
```

10. Common fixes and expected outputs
------------------------------------
- Fix: Docker integration — expected `docker version` to show Client and Server; fix by enabling WSL integration.
- Fix: kind cluster missing nodes — `kubectl get nodes` should show at least one `Ready` node.
- Fix: pod image pull error — `kubectl describe pod` shows `ErrImagePull` or `ImagePullBackOff` and `kubectl get events` shows registry errors.

Style and contribution notes
----------------------------
- Keep instructions concise and place commands for execution inside code blocks.
- Reference existing examples in `examples/` rather than duplicating manifests.
