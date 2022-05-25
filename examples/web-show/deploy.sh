#!/usr/bin/env bash
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#



set -e

usage() {
    cat << EOF
This script is used to install web-show.
USAGE:
    install.sh [FLAGS] [OPTIONS]
FLAGS:
    -h, --help              Prints help information
    -d, --delete            Delete web-show application
        --docker-mirror     Use docker mirror to pull image
EOF
}

DOCKER_MIRROR=false
DELETE_APP=false

while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    --docker-mirror)
        DOCKER_MIRROR=true
        shift
        ;;
    -d|--delete)
        DELETE_APP=true
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

if [ ${DELETE_APP} == "true" ]; then
    kubectl delete deployments web-show
    kubectl delete service web-show
    exit 0
fi

TARGET_IP=$(kubectl get pod -n kube-system -o wide| grep kube-controller | head -n 1 | awk '{print $6}')

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: web-show
  labels:
    app: web-show
spec:
  selector:
    app: web-show
  ports:
    - protocol: TCP
      port: 8081
      targetPort: 8081
EOF

cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-show
  labels:
    app: web-show
spec:
  replicas: 1
  selector:
    matchLabels:
      app: web-show
  template:
    metadata:
      labels:
        app: web-show
    spec:
      containers:
        - name: web-show
          image: ghcr.io/chaos-mesh/web-show
          imagePullPolicy: Always
          command:
            - /usr/local/bin/web-show
            - --target-ip=${TARGET_IP}
          ports:
            - name: web-port
              containerPort: 8081
              hostPort: 8081
EOF

while [[ $(kubectl get pods -l app=web-show -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "Waiting for pod running" && sleep 10; done

kill $(lsof -t -i:8081) >/dev/null 2>&1 || true

nohup kubectl port-forward --address 0.0.0.0 svc/web-show 8081:8081 >/dev/null 2>&1 &
