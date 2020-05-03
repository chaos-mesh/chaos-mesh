#!/usr/bin/env bash

kubectl get networkchaos
kubectl annotate networkchaos network-netem-example chaos-mesh/cleanFinalizer=forced
sleep 3
kubectl get networkchaos
