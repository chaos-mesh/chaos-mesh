import { Archive } from './archives.type'
import { AxiosResponse } from 'axios'
import http from './http'

export const archives: (namespace?: string, name?: string, kind?: string) => Promise<AxiosResponse<Archive[]>> = (
  namespace = '',
  name = '',
  kind = ''
) => http.get(`/archives?namespace=${namespace}&name=${name}&kind=${kind}`)
