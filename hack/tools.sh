#!/usr/bin/env bash

# Copyright 2020 PingCAP, Inc.
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

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(unset CDPATH && cd $(dirname "${BASH_SOURCE[0]}")/.. && pwd)
cd $ROOT

source "${ROOT}/hack/lib.sh"

INSTALLS="$1"

if [ "${INSTALLS}" = "all" ] || grep -qw "kubebuilder" <<<"${INSTALLS}"; then
  hack::ensure_kubebuilder
fi

if [ "${INSTALLS}" = "all" ] || grep -qw "kustomize" <<<"${INSTALLS}"; then
  hack::ensure_kustomize
fi

if [ "${INSTALLS}" = "all" ] || grep -qw "kind" <<<"${INSTALLS}"; then
  hack::ensure_kind
fi

if [ "${INSTALLS}" = "all" ] || grep -qw "kubectl" <<<"${INSTALLS}"; then
  hack::ensure_kubectl
fi

if [ "${INSTALLS}" = "all" ] || grep -qw "helm" <<<"${INSTALLS}"; then
  hack::ensure_helm
fi
