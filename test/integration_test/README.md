# Integration Test

## Preparation

1. Install Chaos Mesh

   Refer to <https://chaos-mesh.org/docs/configure-development-environment/> to configure your development environment.

   For bootstrapping, we currently use the following command to suit our test cases:

   ```shell
   helm install chaos-mesh helm/chaos-mesh -n=chaos-mesh --create-namespace --set chaosDaemon.runtime=containerd,chaosDaemon.socketPath=/run/containerd/containerd.sock,controllerManager.leaderElection.enabled=false,controllerManager.chaosdSecurityMode=false
   kubectl wait --timeout=60s --for=condition=Ready pod -n chaos-mesh -l app.kubernetes.io/instance=chaos-mesh
   kubectl port-forward -n chaos-mesh svc/chaos-dashboard 2333:2333
   ```

   You should also check out the file `.github/workflows/integration-test.yaml` to learn how to set up Chaos Mesh, as this document may not be up to date.

2. Install localstack && aws client (optional)

   It is required when run aws test. You can install localstack and aws client by command:

   ```bash
   helm repo add localstack-repo http://helm.localstack.cloud
   helm upgrade --install localstack localstack-repo/localstack
   pip install awscli
   ```

## Run all tests

Executing command below to run all test cases:

```shell
./test/integration_test/run.sh
```

## Run a specified test

Executing command below to run specified test cases:

```shell
./test/integration_test/run.sh ${case_name}
```
