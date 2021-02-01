#!/usr/bin/env bash

# Copyright 2020 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# See the License for the specific language governing permissions and
# limitations under the License.

set -eu

cur=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $cur

echo "deploy localstack as mock aws server"

pip install awscli
helm repo add localstack-repo http://helm.localstack.cloud
helm upgrade --install localstack localstack-repo/localstack

NODE_PORT=$(kubectl get --namespace default -o jsonpath="{.spec.ports[0].nodePort}" services localstack)
NODE_IP=$(kubectl get nodes --namespace default -o jsonpath="{.items[0].status.addresses[0].address}")
LOCALSTACK_SERVER="http:\/\/$NODE_IP\:$NODE_PORT"

# wait pod status to running
for ((k=0; k<30; k++)); do
    kubectl get pods --namespace default > pods.status
    cat pods.status

    run_num=`grep Running pods.status | wc -l`
    pod_num=$((`cat pods.status | wc -l` - 1))
    if [ $run_num == $pod_num ]; then
        break
    fi

    sleep 3
done

kubectl port-forward svc/localstack 4566:4566 &

aws configure set aws_access_key_id test
aws configure set aws_secret_access_key test
aws configure set default.region us-east-1
aws configure set default.output_format text

echo "run ec2 instance, and the state is pending, will switch to running later"
aws --endpoint-url=http://127.0.0.1:4566 ec2 run-instances --image-id ami --count 1 --instance-type t2.micro --key-name test > run_instance.log
check_contains "pending" run_instance.log
INSTANCE_ID=`cat run_instance.log | grep "InstanceId" | sed 's/.*\"InstanceId\": \"\([0-9,a-z,-]*\)\",/\1/g'`

sleep 2

aws --endpoint-url=http://127.0.0.1:4566 ec2 describe-instances --instance-id $INSTANCE_ID > describe_instance.log
check_contains "running" describe_instance.log

echo "apply aws chaos to stop the ec2 instance, and the state shoud be stopped"

cp aws_chaos_template.yaml aws_chaos.yaml
sed -i "s/instance-id-placeholder/$INSTANCE_ID/g" aws_chaos.yaml
sed -i "s/endpoint-placeholder/$LOCALSTACK_SERVER/g" aws_chaos.yaml
cat aws_chaos.yaml

kubectl apply -f aws_secret.yaml
kubectl apply -f aws_chaos.yaml

sleep 2

aws --endpoint-url=http://127.0.0.1:4566 ec2 describe-instances --instance-id $INSTANCE_ID > describe_instance.log
check_contains "stopped" describe_instance.log

# clean
kubectl delete -f aws_chaos.yaml
helm uninstall localstack
