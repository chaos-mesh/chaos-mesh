#!/usr/bin/env bash
# Copyright 2025 Chaos Mesh Authors.
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
# Provision a Keycloak realm/client/user for the OIDC integration test.
# Runs kcadm.sh inside the keycloak pod, and writes the generated client
# secret to the path given by OUTPUT_SECRET_FILE.

set -eu

REALM=${REALM:-kubernetes-oidc}
CLIENT_ID=${CLIENT_ID:-chaos-dashboard}
USERNAME=${USERNAME:-test-user}
PASSWORD=${PASSWORD:-password}
REDIRECT_URIS=${REDIRECT_URIS:-http://localhost:2333/*}
OUTPUT_SECRET_FILE=${OUTPUT_SECRET_FILE:-/tmp/oidc-test/client-secret.txt}

kc() { kubectl exec -i -n keycloak deploy/keycloak -- /opt/keycloak/bin/kcadm.sh "$@"; }

echo "[keycloak] logging in as admin"
kc config credentials --server http://localhost:8080 --realm master --user admin --password admin

echo "[keycloak] creating realm ${REALM}"
kc create realms -s realm="${REALM}" -s enabled=true 2>/dev/null || echo "  realm already exists"

echo "[keycloak] creating user ${USERNAME}"
kc create users -r "${REALM}" \
  -s username="${USERNAME}" -s enabled=true \
  -s firstName=Test -s lastName=User \
  -s email="${USERNAME}@example.com" -s emailVerified=true \
  -s 'requiredActions=[]' 2>/dev/null || echo "  user already exists"
kc set-password -r "${REALM}" --username "${USERNAME}" --new-password "${PASSWORD}"

echo "[keycloak] creating confidential client ${CLIENT_ID}"
kc create clients -r "${REALM}" \
  -s clientId="${CLIENT_ID}" -s enabled=true \
  -s publicClient=false -s standardFlowEnabled=true \
  -s "redirectUris=[\"${REDIRECT_URIS}\"]" \
  -s 'webOrigins=["*"]' 2>/dev/null || echo "  client already exists"

CID=$(kc get clients -r "${REALM}" -q clientId="${CLIENT_ID}" --fields id --format csv | tail -1 | tr -d '"\r')
SECRET=$(kc get "clients/${CID}/client-secret" -r "${REALM}" --fields value --format csv | tail -1 | tr -d '"\r')

mkdir -p "$(dirname "${OUTPUT_SECRET_FILE}")"
printf '%s' "${SECRET}" > "${OUTPUT_SECRET_FILE}"
echo "[keycloak] client secret written to ${OUTPUT_SECRET_FILE}"
echo "[keycloak] done: realm=${REALM} client=${CLIENT_ID} user=${USERNAME}"
