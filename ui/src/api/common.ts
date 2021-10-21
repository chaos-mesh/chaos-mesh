/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import { Config, RBACConfigParams } from './common.type'

import { Scope } from 'components/NewExperiment/types'
import http from './http'

export const config = () => http.get<Config>('/common/config')

export const rbacConfig = ({ namespace, role }: RBACConfigParams) =>
  http.get('/common/rbac-config', {
    params: {
      namespace,
      role,
    },
  })

export const chaosAvailableNamespaces = () => http.get<string[]>('/common/chaos-available-namespaces')

type stringStringArrayMap = Record<string, string[]>

export const labels = (podNamespaceList: string[]) =>
  http.get<stringStringArrayMap>('/common/labels', {
    params: {
      podNamespaceList: podNamespaceList.join(','),
    },
  })

export const annotations = (podNamespaceList: string[]) =>
  http.get<stringStringArrayMap>('/common/annotations', {
    params: {
      podNamespaceList: podNamespaceList.join(','),
    },
  })

export const pods = (data: Partial<Scope['selector']>) => http.post('/common/pods', data)
