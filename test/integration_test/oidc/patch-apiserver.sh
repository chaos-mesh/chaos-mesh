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
# Patch the minikube kube-apiserver static pod to trust the Keycloak OIDC
# issuer, so that id_tokens minted by Keycloak are accepted as Kubernetes
# authentication tokens. Idempotent.
#
# The minikube node is a docker container (named after the profile), so we use
# `docker exec` / `docker cp` instead of `minikube ssh` -- the latter allocates
# a TTY and hangs on piped stdin, which is unreliable in automation.

set -eu

KC_HOST=${KC_HOST:?KC_HOST is required, e.g. keycloak.192.168.49.2.nip.io}
REALM=${REALM:-kubernetes-oidc}
CLIENT_ID=${CLIENT_ID:-chaos-dashboard}
CA_FILE=${CA_FILE:?CA_FILE is required, path to the keycloak CA cert on host}
NODE=${MINIKUBE_NODE:-minikube}
ISSUER="https://${KC_HOST}/realms/${REALM}"
MANIFEST=/etc/kubernetes/manifests/kube-apiserver.yaml

echo "[apiserver] copying CA to node ${NODE}"
docker exec "${NODE}" mkdir -p /etc/kubernetes/keycloak
docker cp "${CA_FILE}" "${NODE}:/etc/kubernetes/keycloak/ca.crt"

TMP=$(mktemp -d)
docker exec "${NODE}" cat "${MANIFEST}" | tr -d '\r' > "${TMP}/kube-apiserver.yaml"

if grep -q -- "--oidc-issuer-url=${ISSUER}" "${TMP}/kube-apiserver.yaml"; then
  echo "[apiserver] already patched, skipping"
  exit 0
fi

echo "[apiserver] backing up manifest to /etc/kubernetes/kube-apiserver.yaml.bak"
docker exec "${NODE}" cp "${MANIFEST}" /etc/kubernetes/kube-apiserver.yaml.bak

echo "[apiserver] patching manifest with OIDC flags + CA volume"
yq -i ".spec.containers[0].command += [\"--oidc-issuer-url=${ISSUER}\", \"--oidc-client-id=${CLIENT_ID}\", \"--oidc-ca-file=/keycloak-ca.crt\"]" "${TMP}/kube-apiserver.yaml"
yq -i '.spec.volumes += [{"name": "keycloak-ca", "hostPath": {"path": "/etc/kubernetes/keycloak", "type": "DirectoryOrCreate"}}]' "${TMP}/kube-apiserver.yaml"
yq -i '.spec.containers[0].volumeMounts += [{"name": "keycloak-ca", "mountPath": "/keycloak-ca.crt", "subPath": "ca.crt", "readOnly": true}]' "${TMP}/kube-apiserver.yaml"

# Write back atomically: copy into the node's /tmp, then mv onto the manifest
# path so kubelet never observes a partially written file.
docker cp "${TMP}/kube-apiserver.yaml" "${NODE}:/etc/kubernetes/kube-apiserver-oidc.yaml"
docker exec "${NODE}" mv /etc/kubernetes/kube-apiserver-oidc.yaml "${MANIFEST}"
echo "[apiserver] manifest written, waiting for kube-apiserver to restart"

# kubelet picks up the manifest change and recreates the static pod; the API is
# briefly unavailable. First wait for it to go down, then come back healthy.
sleep 10
for i in $(seq 1 60); do
  if kubectl get --raw /healthz >/dev/null 2>&1; then
    echo "[apiserver] healthy again after ~$((10 + i * 3))s"
    docker exec "${NODE}" grep -o -- '--oidc[^ ]*' "${MANIFEST}" | sed 's/^/[apiserver]   flag: /'
    exit 0
  fi
  sleep 3
done

echo "[apiserver] ERROR: kube-apiserver did not become healthy in time" >&2
exit 1
