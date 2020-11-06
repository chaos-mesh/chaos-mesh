import { ExperimentScope } from 'components/NewExperiment/types'
import http from './http'

export const chaosAvailableNamespaces = () => http.get('/common/chaos-available-namespaces')

export const labels = (podNamespaceList: string[]) =>
  http.get(`/common/labels?podNamespaceList=${podNamespaceList.join(',')}`)

export const annotations = (podNamespaceList: string[]) =>
  http.get(`/common/annotations?podNamespaceList=${podNamespaceList.join(',')}`)

export const pods = (data: Partial<Omit<ExperimentScope, 'mode' | 'value'>>) => http.post(`/common/pods`, data)
