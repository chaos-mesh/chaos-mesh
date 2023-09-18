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

function check_chaosd_health() {
    success=false
    for ((k=0; k<30; k++)); do
        status=`curl -w '%{http_code}' -s -o /dev/null $localIP:31768/api/system/health`
        if [ ${status} = 200 ];then
            success=true
            break
        fi
        sleep 1
    done

    if [ "$success" = false ];then
        echo "chaosd starts failed!"
        exit 1
    fi
    echo "chaosd starts succeed!"
}

cur=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $cur

echo "download and deploy chaosd"
localIP=`ifconfig | grep -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1' | head -1`

CHAOSD_VERSION=v1.4.0
curl -fsSLO https://mirrors.chaos-mesh.org/chaosd-${CHAOSD_VERSION}-linux-amd64.tar.gz
tar zxvf chaosd-${CHAOSD_VERSION}-linux-amd64.tar.gz
./chaosd-${CHAOSD_VERSION}-linux-amd64/chaosd server --port 31768 > chaosd.log 2>&1 &
check_chaosd_health

function judge_stress() {
    have_stress=$1
    success=false

    for ((time=0; time<30; time++)); do
        # We expect that there are two `memStress` processes running at the same time (one of them is `grep`) when StressChaos is injected.
        mem_stress_processes=`ps aux | grep "memStress" | wc -l`

        if [ "$have_stress" = true ]; then
            if [ $mem_stress_processes -lt 2 ]; then
                echo "memStress is not run when creating stress chaos on physical machine"
            else
                success=true
                break
            fi
        else
            if [ $mem_stress_processes -gt 1 ]; then
                echo "memStress is still running when delete stress chaos on physical machine"
            else
                success=true
                break
            fi
        fi

        sleep 1
    done

    if [ "$success" = false ]; then
        echo "[debug] chaos-controller-manager log:"
        kubectl logs -n chaos-mesh -l app.kubernetes.io/component=controller-manager --tail=30
        echo
        echo "[debug] chaosd log:"
        tail chaosd.log

        exit 1
    fi
}

echo "create physical machine chaos with address"
cp physical-stress-mem.yaml physical-stress-mem_tmp.yaml
sed -i "s/CHAOSD_ADDRESS/$localIP\:31768/g" physical-stress-mem_tmp.yaml
kubectl apply -f physical-stress-mem_tmp.yaml
judge_stress true

kubectl delete -f physical-stress-mem_tmp.yaml
judge_stress false

echo "create physical machine"
cp physical-machine.yaml physical-machine_tmp.yaml
sed -i "s/CHAOSD_ADDRESS/$localIP\:31768/g" physical-machine_tmp.yaml
kubectl apply -f physical-machine_tmp.yaml

echo "create physical machine chaos with selector"
kubectl apply -f physical-stress-mem-selector.yaml
judge_stress true

kubectl delete -f physical-stress-mem-selector.yaml
judge_stress false

echo "create physical machine schedule"
kubectl apply -f schedule.yaml
judge_stress true

kubectl delete -f schedule.yaml
judge_stress false

echo "create workflow include physical machine chaos"
kubectl apply -f workflow.yaml
judge_stress true

kubectl delete -f workflow.yaml
judge_stress false


echo "****** finish physical machine chaos test ******"
# clean
rm chaosd-${CHAOSD_VERSION}-linux-amd64.tar.gz
rm -rf chaosd-${CHAOSD_VERSION}-linux-amd64
rm *_tmp.yaml
rm chaosd.log
killall chaosd

cd -
