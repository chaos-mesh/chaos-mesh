import { Config, RBACConfigParams } from './common.type'

import { ExperimentScope } from 'components/NewExperiment/types'
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

export const pods = (data: Partial<Omit<ExperimentScope, 'mode' | 'value'>>) => http.post('/common/pods', data)
