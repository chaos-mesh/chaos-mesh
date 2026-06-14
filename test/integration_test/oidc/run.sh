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
# End-to-end integration test for OIDC authentication of the Chaos Dashboard.
#
# It provisions a real Keycloak as the OIDC provider, configures the minikube
# kube-apiserver to trust it, installs Chaos Mesh with oidcSecurityMode enabled,
# then drives the full authorization-code login flow through the dashboard and
# asserts that the resulting id_token is honored by Kubernetes RBAC.
#
# Prerequisites (provided by the CI workflow, or set up manually for local runs):
#   - a running minikube cluster using the docker driver
#   - chaos-mesh / chaos-daemon / chaos-dashboard images loaded into the node,
#     where the dashboard image contains the OIDC code under test
#   - kubectl / helm / yq / openssl / curl / python3 on PATH

set -eu

cur=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
cd "$cur"
repo_root=$(cd "$cur/../../.." && pwd)

# ---- configuration (override via environment) ----------------------------
NS_CHAOS=${NS_CHAOS:-chaos-mesh}
NS_KEYCLOAK=${NS_KEYCLOAK:-keycloak}
REALM=${REALM:-kubernetes-oidc}
CLIENT_ID=${CLIENT_ID:-chaos-dashboard}
USERNAME=${USERNAME:-test-user}
PASSWORD=${PASSWORD:-password}
# global tag for chaos-mesh / chaos-daemon, dashboard tag may differ locally
IMAGE_TAG=${IMAGE_TAG:-latest}
DASHBOARD_IMAGE_TAG=${DASHBOARD_IMAGE_TAG:-$IMAGE_TAG}
DASHBOARD_URL=${DASHBOARD_URL:-http://localhost:2333}

MIP=$(minikube ip)
export KC_HOST="keycloak.${MIP}.nip.io"
ISSUER="https://${KC_HOST}/realms/${REALM}"

WORKDIR=$(mktemp -d)
CA_FILE="${WORKDIR}/tls.crt"
KEY_FILE="${WORKDIR}/tls.key"
SECRET_FILE="${WORKDIR}/client-secret.txt"

PF_PID=""
cleanup() {
  [ -n "${PF_PID}" ] && kill "${PF_PID}" 2>/dev/null || true
  rm -rf "${WORKDIR}"
}
trap cleanup EXIT

# ---- assertion helpers ---------------------------------------------------
fail() { echo "FAIL: $*" >&2; exit 1; }

assert_eq() {
  # assert_eq <expected> <actual> <message>
  [ "$1" = "$2" ] || fail "$3 (expected '$1', got '$2')"
  echo "  OK: $3"
}

assert_contains() {
  # assert_contains <needle> <file> <message>
  grep -q -- "$1" "$2" || fail "$3 (did not find '$1' in $(cat "$2"))"
  echo "  OK: $3"
}

# ---- step 1: self-signed certificate for Keycloak ------------------------
echo "=== [1/7] generating self-signed certificate for ${KC_HOST} ==="
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout "${KEY_FILE}" -out "${CA_FILE}" -days 365 \
  -subj "/CN=${KC_HOST}" -addext "subjectAltName=DNS:${KC_HOST}" 2>/dev/null

# ---- step 2: ingress + Keycloak deployment -------------------------------
echo "=== [2/7] deploying Keycloak ==="
minikube addons enable ingress >/dev/null
kubectl wait -n ingress-nginx --for=condition=ready pod \
  -l app.kubernetes.io/component=controller --timeout=180s >/dev/null

kubectl create namespace "${NS_KEYCLOAK}" --dry-run=client -o yaml | kubectl apply -f - >/dev/null
kubectl create secret tls keycloak-tls -n "${NS_KEYCLOAK}" \
  --cert="${CA_FILE}" --key="${KEY_FILE}" \
  --dry-run=client -o yaml | kubectl apply -f - >/dev/null
KC_HOST="${KC_HOST}" envsubst '${KC_HOST}' < manifests/keycloak.yaml | kubectl apply -f - >/dev/null
kubectl wait -n "${NS_KEYCLOAK}" --for=condition=available --timeout=240s deployment/keycloak

echo "    waiting for Keycloak discovery endpoint via ingress"
for i in $(seq 1 40); do
  if curl -sf --cacert "${CA_FILE}" "https://${KC_HOST}/realms/master" >/dev/null 2>&1; then break; fi
  sleep 3
  [ "$i" = 40 ] && fail "Keycloak discovery endpoint not reachable"
done

# ---- step 3: configure realm / client / user -----------------------------
echo "=== [3/7] configuring Keycloak realm/client/user ==="
REALM="${REALM}" CLIENT_ID="${CLIENT_ID}" USERNAME="${USERNAME}" PASSWORD="${PASSWORD}" \
  OUTPUT_SECRET_FILE="${SECRET_FILE}" bash setup-keycloak.sh
CLIENT_SECRET=$(cat "${SECRET_FILE}")

# ---- step 4: patch kube-apiserver to trust Keycloak ----------------------
echo "=== [4/7] patching kube-apiserver for OIDC ==="
KC_HOST="${KC_HOST}" REALM="${REALM}" CLIENT_ID="${CLIENT_ID}" CA_FILE="${CA_FILE}" \
  bash patch-apiserver.sh

# ---- step 5: install Chaos Mesh with oidcSecurityMode --------------------
echo "=== [5/7] installing Chaos Mesh with oidcSecurityMode ==="
helm upgrade --install chaos-mesh "${repo_root}/helm/chaos-mesh" \
  -n "${NS_CHAOS}" --create-namespace --wait --timeout 5m \
  --set images.tag="${IMAGE_TAG}" \
  --set controllerManager.chaosdSecurityMode=false \
  --set controllerManager.leaderElection.enabled=false \
  --set dashboard.image.tag="${DASHBOARD_IMAGE_TAG}" \
  --set dashboard.oidcSecurityMode.enabled=true \
  --set dashboard.oidcSecurityMode.clientId="${CLIENT_ID}" \
  --set dashboard.oidcSecurityMode.clientSecret="${CLIENT_SECRET}" \
  --set dashboard.oidcSecurityMode.issuerUrl="${ISSUER}" \
  --set-file dashboard.oidcSecurityMode.caBundlePEM="${CA_FILE}"

kubectl wait -n "${NS_CHAOS}" --for=condition=available --timeout=180s deployment/chaos-dashboard

kubectl -n "${NS_CHAOS}" port-forward svc/chaos-dashboard 2333:2333 >/dev/null 2>&1 &
PF_PID=$!
# Wait until the dashboard can actually reach Keycloak: the /redirect handler
# fetches the OIDC discovery document on each call, so a 302 here means the
# dashboard -> Keycloak path (DNS, ingress, CA trust) is ready end to end.
echo "    waiting for dashboard OIDC redirect to be ready"
for i in $(seq 1 30); do
  rc=$(curl -s -o /dev/null -w '%{http_code}' "${DASHBOARD_URL}/api/auth/oidc/redirect" || true)
  [ "${rc}" = "302" ] && break
  sleep 2
  [ "$i" = 30 ] && fail "dashboard OIDC redirect not ready (last http=${rc})"
done

# ---- step 6: drive the OIDC authorization-code login flow ----------------
# Echoes the access_token (id_token) the dashboard stored in its cookie.
oidc_login() {
  local kc_cookie dash_cookie auth_url form_action callback
  kc_cookie=$(mktemp); dash_cookie=$(mktemp)

  # dashboard /redirect -> Keycloak authorize URL
  auth_url=$(curl -s -o /dev/null -D - "${DASHBOARD_URL}/api/auth/oidc/redirect" \
    | grep -i '^location:' | sed 's/location: //I' | tr -d '\r')
  [ -n "${auth_url}" ] || { echo "ERR: no redirect from dashboard" >&2; return 1; }

  # Keycloak login page -> parse the login form action
  form_action=$(curl -sf --cacert "${CA_FILE}" -c "${kc_cookie}" "${auth_url}" \
    | grep -oE 'action="[^"]*"' | head -1 | sed 's/action="//;s/"$//;s/&amp;/\&/g')
  [ -n "${form_action}" ] || { echo "ERR: failed to parse Keycloak login form" >&2; return 1; }

  # submit credentials -> Keycloak redirects back to dashboard /callback with code
  callback=$(curl -s --cacert "${CA_FILE}" -b "${kc_cookie}" -c "${kc_cookie}" -o /dev/null -D - \
    --data-urlencode "username=${USERNAME}" --data-urlencode "password=${PASSWORD}" \
    "${form_action}" | grep -i '^location:' | sed 's/location: //I' | tr -d '\r')
  case "${callback}" in
    *"/api/auth/oidc/callback?"*code=*) : ;;
    *) echo "ERR: no auth code in Keycloak redirect: ${callback}" >&2; return 1 ;;
  esac

  # dashboard /callback exchanges the code and stores tokens in cookies
  curl -s --cookie-jar "${dash_cookie}" -o /dev/null "${callback}"
  grep -E 'access_token' "${dash_cookie}" | awk '{print $7}'
}

echo "=== [6/7] performing OIDC login through the dashboard ==="
ACCESS_TOKEN=$(oidc_login)
[ -n "${ACCESS_TOKEN}" ] || fail "OIDC login did not yield an access_token"
echo "  OK: obtained id_token from dashboard callback (len=${#ACCESS_TOKEN})"

# the k8s identity is "<issuer>#<sub>" (default oidc-username-claim=sub)
SUB=$(python3 -c "import sys,json,base64; p=sys.argv[1].split('.')[1]; p+='='*(-len(p)%4); print(json.loads(base64.urlsafe_b64decode(p))['sub'])" "${ACCESS_TOKEN}")
USER_IDENTITY="${ISSUER}#${SUB}"

# ---- step 7: assert Kubernetes RBAC honors the OIDC identity -------------
echo "=== [7/7] verifying RBAC against the OIDC identity ==="

echo "  - without a binding, the dashboard API must reject the request"
kubectl delete clusterrolebinding oidc-test-admin --ignore-not-found >/dev/null
sleep 1
code=$(curl -s -o "${WORKDIR}/denied.out" -w '%{http_code}' \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" "${DASHBOARD_URL}/api/experiments")
assert_eq "401" "${code}" "unprivileged OIDC user is rejected"
assert_contains "no_cluster_privilege" "${WORKDIR}/denied.out" "rejection cites missing privilege"

echo "  - after binding cluster-admin, the same token must be accepted"
kubectl create clusterrolebinding oidc-test-admin \
  --clusterrole=cluster-admin --user="${USER_IDENTITY}" >/dev/null
sleep 1
code=$(curl -s -o "${WORKDIR}/allowed.out" -w '%{http_code}' \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" "${DASHBOARD_URL}/api/experiments")
assert_eq "200" "${code}" "privileged OIDC user is accepted"

kubectl delete clusterrolebinding oidc-test-admin --ignore-not-found >/dev/null

echo "pass the oidc integration test!"
