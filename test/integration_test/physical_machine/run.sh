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

echo "download and deploy chaosd"

# TODO: use a released version
curl -fsSL -o chaosd-platform-linux-amd64.tar.gz https://mirrors.chaos-mesh.org/chaosd-platform-linux-amd64.tar.gz
tar zxvf chaosd-platform-linux-amd64.tar.gz
./chaosd-platform-linux-amd64/chaosd server --port 31768 > chaosd.log 2>&1 &

function judge_stress() {
    hava_stress=$1
    success=false
    for ((k=0; k<10; k++)); do
        stress_ng_num=`ps aux > test.temp && grep "stress-ng" test.temp | wc -l && rm test.temp`
        if [ "$hava_stress" = true ]; then
            if [ ${stress_ng_num} -lt 1 ]; then
                echo "stress-ng is not run when creating stress chaos on physical machine"
            else
                success=true
                break
            fi
        else
            if [ ${stress_ng_num} -gt 0 ]; then
                echo "stress-ng is still running when delete stress chaos on physical machine"
            else
                success=true
                break
            fi
        fi

        sleep 1
    done

    if [ "$success" = false ]; then
        exit 1
    fi
}

echo "create physical machine chaos"
localIP=`ifconfig | grep -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1' | head -1`

cp chaos.yaml chaos_tmp.yaml
sed -i 's/CHAOSD_ADDRESS/'$localIP'\:31768/g' chaos_tmp.yaml
kubectl apply -f chaos_tmp.yaml
judge_stress true

kubectl delete -f chaos_tmp.yaml
judge_stress false

echo "create physical machine schedule"
cp schedule.yaml schedule_tmp.yaml
sed -i 's/CHAOSD_ADDRESS/'$localIP'\:31768/g' schedule_tmp.yaml
kubectl apply -f schedule_tmp.yaml
judge_stress true

kubectl delete -f schedule_tmp.yaml
judge_stress false

echo "create workflow include physical machine chaos"
cp workflow.yaml workflow_tmp.yaml
sed -i 's/CHAOSD_ADDRESS/'$localIP'\:31768/g' workflow_tmp.yaml
kubectl apply -f workflow_tmp.yaml
judge_stress true

kubectl delete -f workflow_tmp.yaml
judge_stress false


echo "****** finish physical machine chaos test ******"
# clean
rm chaosd-v1.0.1-linux-amd64.tar.gz
rm -rf chaosd-v1.0.1-linux-amd64
rm *_tmp.yaml
rm chaosd.log
killall chaosd

cd -
