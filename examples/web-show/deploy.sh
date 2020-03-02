#!/usr/bin/env bash

set -e

TARGET_IP=$(kubectl get pod -n kube-system -o wide| grep kube-controller | head -n 1 | awk '{print $6}')

sed "s/TARGETIP/$TARGET_IP/g" deployment.yaml > deployment-target.yaml

docker pull 
kubectl apply -f service.yaml
kubectl apply -f deployment-target.yaml

rm -rf deployment-target.yaml

while [[ $(kubectl get pods -l app=web-show -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "Waiting for pod running" && sleep 1; done

kill $(lsof -t -i:8081) 2>&1 >/dev/null | True

nohup kubectl port-forward svc/web-show 8081:8081 >/dev/null 2>&1 &
