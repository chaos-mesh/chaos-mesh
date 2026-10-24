#!/bin/sh
# Copyright 2026 Chaos Mesh Authors.
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
set -eu
[ "$$" -eq 1 ]
[ "$(uname -r)" = 6.8.0-cm-kernel-ci ]
[ -f /etc/cm-lifecycle/guest-id ]
[ -f /etc/cm-lifecycle/node.json ]
mount -t proc -o nosuid,nodev,noexec proc /proc
identity=$(cat /etc/cm-lifecycle/guest-id)
[ "${#identity}" -eq 64 ]
case "$identity" in *[!0-9a-f]*) exit 1 ;; esac
actual=$(sha256sum /etc/cm-lifecycle/node.json)
[ "${actual%% *}" = "$identity" ]
case " $(cat /proc/cmdline) " in *" cm_lifecycle=$identity "*) ;; *) exit 1 ;; esac
case " $(cat /proc/cmdline) " in *" root=/dev/vda "*) ;; *) exit 1 ;; esac
[ -z "$(cat /proc/modules)" ]
printf '\n' > /proc/sys/kernel/modprobe
printf '1\n' > /proc/sys/kernel/modules_disabled
[ -z "$(cat /proc/sys/kernel/modprobe)" ]
[ "$(cat /proc/sys/kernel/modules_disabled)" = 1 ]
printf 'CM_LIFECYCLE_MODULE_LOCKED\n'
exec /sbin/init
