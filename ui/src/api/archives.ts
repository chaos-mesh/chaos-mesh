import { Archive } from './archives.type'
import { AxiosResponse } from 'axios'
import http from './http'

export const archives: (namespace?: string, name?: string, kind?: string) => Promise<AxiosResponse<Archive[]>> = (
  namespace = '',
  name = '',
  kind = ''
) => http.get(`/archives?namespace=${namespace}&name=${name}&kind=${kind}`)

export const detail = (uuid: uuid) => http.get(`/archives/detail?uid=${uuid}`)

export const report = (uuid: uuid) => http.get(`/archives/report?uid=${uuid}`)
