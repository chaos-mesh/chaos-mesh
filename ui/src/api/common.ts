import { Config } from './common.type'
import { ExperimentScope } from 'components/NewExperiment/types'
import http from './http'

export const config = () => http.get<Config>('/common/config')

export const chaosAvailableNamespaces = () => http.get<string[]>('/common/chaos-available-namespaces')

export const labels = (podNamespaceList: string[]) =>
  http.get<Record<string, string[]>>(`/common/labels?podNamespaceList=${podNamespaceList.join(',')}`)

export const annotations = (podNamespaceList: string[]) =>
  http.get<Record<string, string[]>>(`/common/annotations?podNamespaceList=${podNamespaceList.join(',')}`)

export const pods = (data: Partial<Omit<ExperimentScope, 'mode' | 'value'>>) => http.post(`/common/pods`, data)
