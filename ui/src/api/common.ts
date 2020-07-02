import http from './http'

export const namespaces = () => http.get('/common/namespaces')

export const labels = (podNamespaceList: string) => http.get(`/common/labels?podNamespaceList=${podNamespaceList}`)

export const annotations = (podNamespaceList: string) =>
  http.get(`/common/annotations?podNamespaceList=${podNamespaceList}`)
