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

kubectl apply -f system-under-test.yaml
sleep 3
kubectl apply -f chaos.yaml
kubectl apply -f block-clean-finalizer.yaml
sleep 3
kubectl get pods
echo "Next step will hang on uncleanable finalizers. ^C to abort."
kubectl delete --force --grace-period=0 -f chaos.yaml
