#!/usr/bin/env bash

# Copyright 2020 PingCAP, Inc.
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

set -e

usage() {
    cat << EOF
This script is used to install web-show.
USAGE:
    install.sh [FLAGS] [OPTIONS]
FLAGS:
    -h, --help              Prints help information
        --docker-mirror     Use docker mirror to pull image
EOF
}

DOCKER_MIRROR=false

while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    --docker-mirror)
        DOCKER_MIRROR=true
        shift
        ;;
    -h|--help)
        usage
        exit 0
        ;;
    *)
        echo "unknown option: $key"
        usage
        exit 1
        ;;
esac
done

TARGET_IP=$(kubectl get pod -n kube-system -o wide| grep kube-controller | head -n 1 | awk '{print $6}')

sed "s/TARGETIP/$TARGET_IP/g" deployment.yaml > deployment-target.yaml

if [ ${DOCKER_MIRROR} == "true" ]; then
    docker pull dockerhub.azk8s.cn/pingcap/web-show || true
    docker tag dockerhub.azk8s.cn/pingcap/web-show pingcap/web-show  || true
    kind load docker-image pingcap/web-show > /dev/null 2>&1 || true
fi

kubectl apply -f service.yaml
kubectl apply -f deployment-target.yaml

rm -rf deployment-target.yaml

while [[ $(kubectl get pods -l app=web-show -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "Waiting for pod running" && sleep 10; done

kill $(lsof -t -i:8081) >/dev/null 2>&1 || true

nohup kubectl port-forward --address 0.0.0.0 svc/web-show 8081:8081 >/dev/null 2>&1 &
