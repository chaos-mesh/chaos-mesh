import { Archive } from './archives.type'
import http from './http'

export const archives = (namespace = '', name = '', kind = '') =>
  http.get<Archive[]>(`/archives?namespace=${namespace}&name=${name}&kind=${kind}`)

export const detail = (uuid: uuid) => http.get(`/archives/detail?uid=${uuid}`)

export const report = (uuid: uuid) => http.get(`/archives/report?uid=${uuid}`)
