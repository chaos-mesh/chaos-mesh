## Integration Test

### Preparation

1. Install Chaos Mesh

You can install Chaos Mesh by commands:

```bash
hack/local-up-chaos-mesh.sh
kubectl set env deployment/chaos-dashboard SECURITY_MODE=true -n chaos-mesh
kubectl port-forward -n chaos-mesh svc/chaos-dashboard 2333:2333 &
```

2. Install localstack && aws client(optional)

It is required when run aws test. You can install localstack and aws client by command:

```bash
helm repo add localstack-repo http://helm.localstack.cloud
helm upgrade --install localstack localstack-repo/localstack
pip install awscli
```

### Run all tests

Executing command below to run all test cases:

```shell
./test/integration_test/run.sh
```

### Run a specified test

Executing command below to run specified test cases:

```shell
./test/integration_test/run.sh ${case_name}
```
