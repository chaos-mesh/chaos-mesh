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

cur=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $cur

echo "deploy a helloword pod which is implement with java"

# source code: https://github.com/WangXiangUSTC/byteman-example/tree/main/example.helloworld
# this application will print log like this below:
# 0. Hello World
# 1. Hello World
# ...

kubectl apply -f ./helloworld_pod.yaml

echo "wait helloworld pod status to running"
for ((k=0; k<30; k++)); do
    kubectl get pods --namespace helloworld > pods.status
    cat pods.status

    run_num=`grep Running pods.status | wc -l`
    pod_num=$((`cat pods.status | wc -l` - 1))
    if [ $run_num == $pod_num ]; then
        break
    fi

    sleep 1
done

function check_log() {
    match=""
    if [ "$1" = "false" ]; then
        match="-v"
    fi

    success=false
    for ((k=0; k<10; k++)); do
        kubectl logs --tail=1 helloworld -n helloworld | grep $match "9999. Hello World"
        if [ "$?" = "0" ]; then
            success=true
            break
        fi

        sleep 2
    done

    if [ "$success" = false ]; then
        exit 1
    fi
}

echo "create jvm chaos, and will print 9999. Hello World"
kubectl apply -f ./jvm.yaml
check_log true

echo "delete jvm chaos, and will not print 9999. Hello World"
kubectl delete -f ./jvm.yaml

check_log false

echo "****** finish jvm chaos test ******"
# clean
kubectl delete -f ./helloworld_pod.yaml
