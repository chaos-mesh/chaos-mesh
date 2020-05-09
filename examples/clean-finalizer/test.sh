#!/usr/bin/env bash

kubectl get networkchaos
kubectl annotate networkchaos network-netem-example chaos-mesh.pingcap.com/cleanFinalizer=forced
sleep 3
kubectl get networkchaos
