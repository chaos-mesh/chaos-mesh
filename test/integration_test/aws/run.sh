#!/usr/bin/env bash

# Copyright 2021 Chaos Mesh Authors.
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

# wait localstash pod status to running
for ((k=0; k<30; k++)); do

    JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{"\n"}{end}'
    not_ready_num=`kubectl get pods -l app.kubernetes.io/name=localstack --no-headers -o jsonpath="$JSONPATH" | grep "Ready=False" | wc -l`

    if [ $not_ready_num == 0 ]; then
        break
    fi

    sleep 3
done

kubectl port-forward svc/localstack 4566:4566 &
# kill child process
trap 'kill $(jobs -p)' EXIT

NODE_PORT=$(kubectl get --namespace default -o jsonpath="{.spec.ports[0].nodePort}" services localstack)
NODE_IP=$(kubectl get nodes --namespace default -o jsonpath="{.items[0].status.addresses[0].address}")
LOCALSTACK_SERVER="http:\/\/$NODE_IP\:$NODE_PORT"

aws configure set aws_access_key_id test
aws configure set aws_secret_access_key test
aws configure set default.region us-east-1
aws configure set default.output_format text

echo "run ec2 instance, and the state is pending, will switch to running later"
aws --endpoint-url=http://127.0.0.1:4566 ec2 run-instances --image-id ami --count 1 --instance-type t2.micro --key-name test > run_instance.log
state=`cat run_instance.log | jq -rM '.Instances[0].State.Name'`
if [ "$state" != "pending" ]; then
    cat run_instance.log
    echo "ec2 instance's state is $state but not pending"
    exit 1
fi

INSTANCE_ID=`cat run_instance.log | jq -rM '.Instances[0].InstanceId'`

sleep 2

aws --endpoint-url=http://127.0.0.1:4566 ec2 describe-instances --instance-id $INSTANCE_ID > describe_instance.log
state=`cat describe_instance.log | jq -rM '.Reservations[0].Instances[0].State.Name'`
if [ "$state" != "running" ]; then
    echo "ec2 instance's state is $state but not running"
    exit 1
fi

echo "apply aws chaos to stop the ec2 instance, and the state shoud be stopped"

cp aws_chaos_template.yaml aws_chaos.yaml
sed -i "s/instance-id-placeholder/$INSTANCE_ID/g" aws_chaos.yaml
sed -i "s/endpoint-placeholder/$LOCALSTACK_SERVER/g" aws_chaos.yaml
cat aws_chaos.yaml

kubectl apply -f aws_secret.yaml
kubectl apply -f aws_chaos.yaml

sleep 2

aws --endpoint-url=http://127.0.0.1:4566 ec2 describe-instances --instance-id $INSTANCE_ID > describe_instance.log
state=`cat describe_instance.log | jq -rM '.Reservations[0].Instances[0].State.Name'`
if [ "$state" != "stopped" ]; then
    echo "ec2 instance's state is $state but not stopped"
    exit 1
fi

# clean
kubectl delete -f aws_chaos.yaml
helm uninstall localstack
