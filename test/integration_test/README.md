# Integration Test

## Preparation

Run commands from the repository root. Use a disposable Linux test host and Kubernetes cluster because these tests inject faults and change configuration. The physical-machine case starts `chaosd` on the host and ends by running `killall chaosd`.

The complete suite requires a minikube cluster with the Docker driver, and locally built Chaos Mesh images loaded into its nodes. Install `minikube`, `docker`, `kubectl`, `helm`, `yq` v4, `openssl`, `curl`, `python3`, `envsubst`, `jq`, `ifconfig`, and `killall`. The physical-machine case runs an amd64 Linux `chaosd` binary and is skipped on aarch64.

The `oidc` case deploys Keycloak, patches the minikube API server's OIDC settings, and reinstalls the dashboard with OIDC authentication. Run it on a separate disposable cluster when testing other cases independently. See [the OIDC CI workflow](../../.github/workflows/integration_test_oidc.yml) for image build and loading steps.

1. Install Chaos Mesh

   You can install Chaos Mesh by commands:

   ```bash
   helm install --wait --create-namespace chaos-mesh helm/chaos-mesh -n=chaos-mesh \
     --set images.tag=latest,controllerManager.chaosdSecurityMode=false
   # Forward the dashboard port
   kubectl port-forward -n chaos-mesh svc/chaos-dashboard 2333:2333 &
   ```

2. Install localstack & AWS client (optional)

   LocalStack and the AWS CLI are required for the `aws` case, including a complete suite run. Install them with:

   ```bash
   helm repo add localstack-repo http://helm.localstack.cloud
   helm upgrade --install localstack localstack-repo/localstack --version 0.6.14
   pip install awscli
   ```

   The `aws` case runs `aws configure set`, which would overwrite your default AWS CLI credentials and configuration. Before running this case or the complete suite, use temporary files in the same shell:

   ```bash
   AWS_TEST_CONFIG_DIR=$(mktemp -d)
   export AWS_CONFIG_FILE="${AWS_TEST_CONFIG_DIR}/config"
   export AWS_SHARED_CREDENTIALS_FILE="${AWS_TEST_CONFIG_DIR}/credentials"
   ```

## Run all tests

After preparing all case dependencies above, run the complete suite. It includes the OIDC case and changes dashboard authentication:

```bash
./test/integration_test/run.sh
```

## Run a specified test

To run one case, replace `${case_name}` with its directory name, such as `network` or `statuscheck`. Prepare only that case's dependencies:

```bash
./test/integration_test/run.sh ${case_name}
```
