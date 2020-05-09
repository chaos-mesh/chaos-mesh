#!/usr/bin/env bash

kubectl apply -f system-under-test.yaml
sleep 3
kubectl apply -f chaos.yaml
kubectl apply -f block-clean-finalizer.yaml
sleep 3
kubectl get pods
echo "Next step will hang on uncleanable finalizers. ^C to abort."
kubectl delete --force --grace-period=0 -f chaos.yaml
