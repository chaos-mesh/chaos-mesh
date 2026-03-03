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

cur=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $cur

echo "deploy busybox for test"
kubectl apply -f https://raw.githubusercontent.com/chaos-mesh/apps/master/ping/busybox-statefulset.yaml

echo "wait pod status to running"
for ((k=0; k<30; k++)); do
    kubectl get pods --namespace busybox > pods.status
    cat pods.status

    run_num=`grep Running pods.status | wc -l`
    pod_num=$((`cat pods.status | wc -l` - 1))
    if [ $run_num == $pod_num ]; then
        break
    fi

    sleep 1
done

echo "****** test delay chaos ******"
kubectl apply -f ./delay_chaos.yaml

echo "verification"
kubectl exec busybox-0 -i -n busybox -- ping -c 10 busybox-1.busybox.busybox.svc > ping.log
cat ping.log

# the log looks like `64 bytes from 10.244.0.9: seq=0 ttl=63 time=0.240 ms`
# get the latency from log
latencies=`cat ping.log | grep "bytes from"  | sed 's/.*time=\([0-9]*\).[0-9]* ms.*/\1/g'`

high_latency_num=0
for latency in $latencies
do
    if [[ "$latency" -ge "10" ]]; then
        high_latency_num=`expr $high_latency_num + 1`
    fi
done

# about half of the latency should be greater than 10ms
if [[ "$high_latency_num" -lt "3" ]] || [[ "$high_latency_num" -gt "6" ]]; then
    echo "the chaos dosen't work as expect"
    exit 1
fi

rm ping.log
rm pods.status

echo "****** finish delay chaos test ******"
kubectl delete -f ./delay_chaos.yaml
cd -
