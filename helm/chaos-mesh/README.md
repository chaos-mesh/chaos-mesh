# Chaos Mesh Helm Chart

[Chaos Mesh](https://chaos-mesh.org) is a cloud-native Chaos Engineering platform for Kubernetes. This directory contains its Chart API v2 package, installable with Helm 3 or later, for deploying the control plane, node daemon, dashboard, admission webhooks, CRDs, and optional supporting components.

For production prerequisites and platform-specific runtime settings, read the [official installation guide](https://chaos-mesh.org/docs/production-installation-using-helm/) before installing the chart.

## Prerequisites

- Helm 3 or later
- A supported Kubernetes cluster and container runtime
- Permission to create cluster-scoped resources, including CRDs, ClusterRoles, ClusterRoleBindings, and admission webhook configurations
- Privileged workloads allowed on the nodes where `chaos-daemon` runs, unless you have deliberately configured a restricted capability set

## Install

### Install a released chart

Choose a version from the [Chaos Mesh Helm repository](https://charts.chaos-mesh.org/) and pin it for reproducible installations:

```bash
helm repo add chaos-mesh https://charts.chaos-mesh.org
helm repo update
helm search repo chaos-mesh/chaos-mesh --versions
helm install chaos-mesh chaos-mesh/chaos-mesh \
  --namespace chaos-mesh \
  --create-namespace \
  --version <version>
```

### Install the chart from this repository

Run the following command from the repository root when developing or testing local chart changes:

```bash
helm upgrade --install chaos-mesh ./helm/chaos-mesh \
  --namespace chaos-mesh \
  --create-namespace
```

The source chart defaults to the `latest` Chaos Mesh image tag. Set `images.tag` or the component-specific image tags when you need a reproducible or locally built image:

```bash
helm upgrade --install chaos-mesh ./helm/chaos-mesh \
  --namespace chaos-mesh \
  --create-namespace \
  --set images.tag=<tag>
```

For non-trivial configuration, keep overrides in a file instead of a long list of `--set` flags:

```bash
helm upgrade --install chaos-mesh ./helm/chaos-mesh \
  --namespace chaos-mesh \
  --create-namespace \
  --values my-values.yaml
```

Verify the installation with:

```bash
kubectl get pods --namespace chaos-mesh \
  --selector app.kubernetes.io/instance=chaos-mesh
```

## Upgrade and uninstall

Review the release notes and CRD changes before upgrading, then reuse the same values file used for installation:

```bash
helm upgrade chaos-mesh chaos-mesh/chaos-mesh \
  --namespace chaos-mesh \
  --version <version> \
  --values my-values.yaml
```

Remove the Helm-managed resources with:

```bash
helm uninstall chaos-mesh --namespace chaos-mesh
```

Helm does not upgrade or delete CRDs placed in the chart's `crds/` directory. See [CRD lifecycle](#crd-lifecycle) before upgrading or removing an installation.

## What the chart installs

| Component | Workload | Default | Purpose |
| --- | --- | --- | --- |
| Controller manager | Deployment | Enabled, 3 replicas | Reconciles Chaos Mesh resources and serves admission webhooks |
| Chaos daemon | DaemonSet | Enabled, privileged | Performs node- and container-level fault injection |
| Dashboard | Deployment | Enabled | Provides the web UI and API |
| DNS server | Deployment | Enabled | Supports DNSChaos |
| Prometheus | Deployment | Disabled | Provides an optional in-chart Prometheus instance |
| BPF kernel helper | DaemonSet sidecar | Disabled | Enables kernel fault injection through `chaos-kernel` |
| Delve sidecars | Sidecars | Disabled | Support remote debugging of Chaos Mesh components |

The chart also creates the required Services, RBAC resources, certificate Secrets or cert-manager resources, admission webhook configurations, and the CRDs stored in [`crds/`](crds/).

## Configuration

[`values.yaml`](values.yaml) is the primary commented configuration reference and the source of defaults. [`values.schema.json`](values.schema.json) is generated from it and is used by Helm to validate value types. Avoid duplicating every nested value in this README; update `values.yaml` and regenerate the schema when adding or changing a value.

The most important top-level settings are summarized below.

| Value | Default | Purpose |
| --- | --- | --- |
| `images.registry` | `ghcr.io` | Global registry used unless a component overrides it |
| `images.tag` | `latest` | Global image tag used unless a component overrides it |
| `imagePullSecrets` | `[]` | Registry credentials added to component Pods |
| `clusterScoped` | `true` | Controls cluster-wide versus namespace-scoped reconciliation and RBAC |
| `rbac.create` | `true` | Creates chart RBAC resources and, subject to component settings, ServiceAccounts |
| `timezone` | `UTC` | Sets the component time zone |
| `enableProfiling` | `true` | Enables profiling endpoints in supported components |
| `extraObjects` | `[]` | Renders additional Kubernetes objects with chart context |

Key component settings:

| Value | Default | Purpose |
| --- | --- | --- |
| `controllerManager.replicaCount` | `3` | Controller manager replicas |
| `controllerManager.imagePullPolicy` | `IfNotPresent` | Controller manager image pull policy |
| `controllerManager.targetNamespace` | `chaos-mesh` | Namespace watched when `clusterScoped=false` |
| `controllerManager.enableFilterNamespace` | `false` | Limits injection to namespaces annotated with `chaos-mesh.org/inject=enabled` |
| `controllerManager.enabledControllers` | `["*"]` | Controllers to start |
| `controllerManager.enabledWebhooks` | `["*"]` | Webhooks to start |
| `controllerManager.leaderElection.enabled` | `true` | Enables controller leader election |
| `controllerManager.localHelmChart.enabled` | `false` | Mounts a local chart for offline multi-cluster installation |
| `chaosDaemon.imagePullPolicy` | `IfNotPresent` | Chaos daemon image pull policy |
| `chaosDaemon.runtime` | `docker` | Container runtime adapter |
| `chaosDaemon.socketPath` | `/var/run/docker.sock` | Host container runtime socket |
| `chaosDaemon.privileged` | `true` | Runs the daemon as a privileged container |
| `chaosDaemon.mtls.enabled` | `true` | Enables mTLS between the controller manager and daemon |
| `chaosDaemon.resourceProfile` | `light` | Resource baseline: `light`, `standard`, or `intensive` |
| `dashboard.create` | `true` | Installs the dashboard |
| `dashboard.securityMode` | `true` | Requires dashboard credentials |
| `dashboard.service.type` | `NodePort` | Dashboard Service type |
| `dashboard.persistentVolume.enabled` | `false` | Persists the default SQLite database |
| `dashboard.gcpSecurityMode.enabled` | `false` | Enables GCP authentication |
| `dashboard.oidcSecurityMode.enabled` | `false` | Enables generic OIDC authentication |
| `dashboard.ingress.enabled` | `false` | Creates a dashboard Ingress |
| `dnsServer.create` | `true` | Installs the DNSChaos server |
| `prometheus.create` | `false` | Installs the bundled Prometheus instance |
| `webhook.certManager.enabled` | `false` | Uses cert-manager for webhook and daemon certificates |
| `bpfki.create` | `false` | Enables the `chaos-kernel` helper |
| `chaosDlv.enable` | `false` | Adds the Delve debugging sidecar |

Component image settings follow the same pattern:

```yaml
images:
  registry: ghcr.io
  tag: latest

controllerManager:
  image:
    registry: "" # Empty means images.registry.
    repository: chaos-mesh/chaos-mesh
    tag: "" # Empty means images.tag.
```

## Common configurations

### Container runtime

The default runtime is Docker. Configure both the runtime and its host socket for containerd or CRI-O:

```yaml
chaosDaemon:
  runtime: containerd
  socketPath: /run/containerd/containerd.sock
```

```yaml
chaosDaemon:
  runtime: crio
  socketPath: /var/run/crio/crio.sock
```

### Namespace-scoped mode

Set `clusterScoped=false` to restrict reconciliation and target bindings to one namespace. CRDs and several control-plane resources remain cluster-scoped and still require cluster-level installation permissions.

The current DNS RBAC template also reads `dnsServer.targetNamespace`, although that value is not yet declared in `values.yaml`. Pass it explicitly and keep it identical to `controllerManager.targetNamespace`:

```yaml
clusterScoped: false

controllerManager:
  targetNamespace: testing

dnsServer:
  targetNamespace: testing
```

To keep cluster-scoped operation while allowing injection only in opted-in namespaces, use `controllerManager.enableFilterNamespace=true` and annotate each allowed namespace:

```bash
kubectl annotate namespace testing chaos-mesh.org/inject=enabled
```

### Chaos daemon resources

`chaosDaemon.resourceProfile` supplies a baseline that can be partially overridden by `chaosDaemon.resources`:

| Profile     | Requests               | Limits                |
| ----------- | ---------------------- | --------------------- |
| `light`     | 100m CPU, 256Mi memory | None                  |
| `standard`  | 250m CPU, 512Mi memory | None                  |
| `intensive` | 500m CPU, 1Gi memory   | 1000m CPU, 2Gi memory |

```yaml
chaosDaemon:
  resourceProfile: standard
  resources:
    limits:
      memory: 1Gi
```

An unsupported non-empty profile causes chart rendering to fail. Set the profile to an empty string if you want `chaosDaemon.resources` to be the only source of resource settings.

### Dashboard persistence and database credentials

The dashboard uses SQLite by default. Enable `dashboard.persistentVolume.enabled` to retain that database across Pod replacement, or configure an external database. Prefer a Secret over putting a database DSN directly in values:

```yaml
dashboard:
  databaseSecretName: chaos-dashboard-database
```

The referenced Secret must contain a `DATABASE_DATASOURCE` key. `dashboard.env.DATABASE_DATASOURCE` remains available for compatibility but is deprecated.

### Dashboard authentication

GCP and generic OIDC authentication are configured under `dashboard.gcpSecurityMode` and `dashboard.oidcSecurityMode`. Each supports inline credentials or an existing Secret. The existing GCP Secret must provide `GCP_CLIENT_ID` and `GCP_CLIENT_SECRET`; the OIDC Secret must provide `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, and `OIDC_ISSUER_URL`.

For an OIDC provider that uses a private CA, set `dashboard.oidcSecurityMode.caBundlePEM`. The chart creates and mounts a CA ConfigMap only when this value is non-empty.

### Certificate management

By default, Helm generates self-signed certificates and stores them in Secrets. You can instead supply `webhook.caBundlePEM`, `webhook.crtPEM`, and `webhook.keyPEM` together.

To delegate certificate creation to cert-manager, install cert-manager and its CRDs first, following the [cert-manager installation guide](https://cert-manager.io/docs/installation/helm/), then set:

```yaml
webhook:
  certManager:
    enabled: true
```

The chart then creates namespaced Issuers and Certificates for the admission webhook and daemon mTLS Secrets. Do not use Helm's `--dry-run` alone to validate this mode because Helm cannot discover CRDs that are not installed in the cluster; use `helm template` with the relevant API version when testing locally.

### Pod Security Policy compatibility

`chaosDaemon.podSecurityPolicy` is a legacy compatibility option and defaults to `false`. Enabling it renders a `policy/v1beta1` PodSecurityPolicy, which is unavailable in modern Kubernetes releases. Prefer the security controls supported by your cluster.

### Extra objects

`extraObjects` accepts arbitrary Kubernetes objects. String values inside each object are evaluated with Helm's `tpl` function, so chart values and release metadata can be referenced:

```yaml
extraObjects:
  - apiVersion: v1
    kind: ConfigMap
    metadata:
      name: "{{ .Release.Name }}-extra"
    data:
      environment: test
```

## CRD lifecycle

The CRDs in [`crds/`](crds/) are installed before chart templates on the first `helm install`. If a CRD already exists, Helm leaves it in place. Helm does not template, upgrade, or delete CRDs from this directory.

Consequences for operators:

- Review and apply changed CRDs separately before upgrading the controller workloads.
- `helm uninstall` intentionally leaves the CRDs and all Chaos Mesh custom resources in the cluster.
- Treat CRD deletion as a separate, destructive operation; deleting a CRD also deletes its custom resources.
- `helm install --skip-crds` is appropriate only when CRDs are managed through another process.

When testing this source tree, update installed CRDs with:

```bash
kubectl apply --filename helm/chaos-mesh/crds/
```

## Developing the chart

Chart files have distinct sources of truth:

- `values.yaml` defines documented defaults and supported values.
- `values.schema.json` is generated from `values.yaml`; do not update it by hand.
- `templates/` contains rendered Kubernetes resources and helpers.
- `crds/` is generated from API definitions under `api/` and `config/crd/bases/`.
- `Chart.yaml` contains chart metadata. Its `version` and `appVersion` are development placeholders in this source tree, not released version numbers.

Run focused checks for the part you changed:

```bash
helm lint helm/chaos-mesh
helm template chaos-mesh helm/chaos-mesh --namespace chaos-mesh
```

Render every conditional path affected by a change. For example:

```bash
helm template chaos-mesh helm/chaos-mesh \
  --namespace chaos-mesh \
  --set clusterScoped=false \
  --set controllerManager.targetNamespace=testing \
  --set dnsServer.targetNamespace=testing

helm template chaos-mesh helm/chaos-mesh \
  --namespace chaos-mesh \
  --set webhook.certManager.enabled=true \
  --api-versions cert-manager.io/v1
```

After changing `values.yaml`, regenerate and inspect the schema:

```bash
make helm-values-schema
git diff -- helm/chaos-mesh/values.schema.json
```

After changing CRD API definitions, run the relevant generation target. `make generate` includes CRD generation and refreshes `config/crd/bases/`, `helm/chaos-mesh/crds/`, and `manifests/crd.yaml`; inspect those outputs together.

The full `make check` target is the final repository-wide verification. During chart development, prefer the focused Helm commands above plus tests for the code you changed, then run `make check` before submitting the final change when practical.
