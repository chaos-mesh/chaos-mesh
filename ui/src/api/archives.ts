import { GetArchives } from './archives.type'
import http from './http'

export const archives: GetArchives = (namespace = '', name = '', kind = '') => {
  return http.get('/archives', {
    params: {
      namespace,
      name,
      kind,
    },
  })
}

export const detail = (uuid: uuid) => http.get(`/archives/detail?uid=${uuid}`)

export const report = (uuid: uuid) => http.get(`/archives/report?uid=${uuid}`)
