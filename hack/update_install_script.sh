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

tmp_file="chaos-mesh.yaml"
tmp_install_script="install.sh.bak"
install_script="install.sh"

helm template chaos-mesh helm/chaos-mesh --namespace=chaos-mesh \
      --set controllerManager.hostNetwork=true \
      --set controllerManager.chaosdSecurityMode=false \
      --set chaosDaemon.hostNetwork=true \
      --set chaosDaemon.mtls.enabled=false \
      --set enableCtrlServer=true \
      --set dashboard.securityMode=false > ${tmp_file}

sed -i.bak '/helm/d' $tmp_file
sed -i.bak '/Helm/d' $tmp_file
sed -i.bak 's/rollme:.*/rollme: \"install.sh\"/g' $tmp_file
sed -i.bak 's/ca.crt:.*/ca.crt: \"\$\{CA_BUNDLE\}\"/g' $tmp_file
sed -i.bak 's/tls.crt:.*/tls.crt: \"\$\{TLS_CRT\}\"/g' $tmp_file
sed -i.bak 's/tls.key:.*/tls.key: \"\$\{TLS_KEY\}\"/g' $tmp_file
sed -i.bak 's/caBundle:.*/caBundle: \"\$\{CA_BUNDLE\}\"/g' $tmp_file
sed -i.bak 's/\/host-run\/docker.sock/\/host-run\/$\{socketName\}/g' $tmp_file
sed -i.bak 's/path: \/var\/run/path: \$\{socketDir\}/g' $tmp_file
sed -i.bak 's/- docker/- $\{runtime\}/g' $tmp_file
sed -i.bak 's/hostNetwork: true/hostNetwork: \$\{host_network\}/g' $tmp_file
sed -i.bak 's/ghcr.io\/chaos-mesh\/chaos-mesh:.*/\${IMAGE_REGISTRY_PREFIX}\/chaos-mesh\/chaos-mesh:\$\{VERSION_TAG\}/g' $tmp_file
sed -i.bak 's/ghcr.io\/chaos-mesh\/chaos-daemon:.*/\${IMAGE_REGISTRY_PREFIX}\/chaos-mesh\/chaos-daemon:\$\{VERSION_TAG\}/g' $tmp_file
sed -i.bak 's/ghcr.io\/chaos-mesh\/chaos-dashboard:.*/\${IMAGE_REGISTRY_PREFIX}\/chaos-mesh\/chaos-dashboard:\$\{VERSION_TAG\}/g' $tmp_file
sed -i.bak 's/value: UTC/value: \$\{timezone\}/g' $tmp_file
sed -i.bak 's/app.kubernetes.io\/version: 0.0.0/app.kubernetes.io\/version: $\{VERSION_TAG##v\}/g' $tmp_file

mv $tmp_file $tmp_file.bak

cat <<EOF > $tmp_file
---
apiVersion: v1
kind: Namespace
metadata:
  name: chaos-mesh
EOF

cat $tmp_file.bak >> $tmp_file

let start_num=$(cat -n $install_script| grep "# chaos-mesh.yaml start" | awk '{print $1}')+1
let end_num=$(cat -n $install_script| grep "# chaos-mesh.yaml end" | awk '{print $1}')-1

head -$start_num $install_script > $tmp_install_script
cat $tmp_file >> $tmp_install_script
tail -n +$end_num $install_script >> $tmp_install_script

mv $tmp_install_script $install_script
chmod +x $install_script

rm -rf $tmp_file
rm -rf $tmp_file.bak
