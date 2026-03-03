# Integration Test

## Preparation

1. Install Chaos Mesh

   You can install Chaos Mesh by commands:

   ```bash
   helm install --wait --create-namespace chaos-mesh helm/chaos-mesh -n=chaos-mesh \
     --set images.tag=latest,controllerManager.chaosdSecurityMode=false
   # Forward the dashboard port
   kubectl port-forward -n chaos-mesh svc/chaos-dashboard 2333:2333 &
   ```

2. Install localstack & AWS client (optional)

   It is required when running AWS tests. You can install localstack and AWS client by command:

   ```bash
   helm repo add localstack-repo http://helm.localstack.cloud
   helm upgrade --install localstack localstack-repo/localstack --version 0.6.14
   pip install awscli
   ```

## Run all tests

Execute the command below to run all test cases:

```bash
./test/integration_test/run.sh
```

## Run a specified test

Execute the command below to run specified test cases:

```bash
./test/integration_test/run.sh ${case_name}
```
