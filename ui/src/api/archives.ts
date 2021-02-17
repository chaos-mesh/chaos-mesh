import { Archive } from './archives.type'
import http from './http'

export const archives = (namespace = null, name = null, kind = null) =>
  http.get<Archive[]>('/archives', {
    params: {
      namespace,
      name,
      kind,
    },
  })

export const detail = (uuid: uuid) =>
  http.get('/archives/detail', {
    params: {
      uid: uuid,
    },
  })

export const report = (uuid: uuid) =>
  http.get('/archives/report', {
    params: {
      uid: uuid,
    },
  })

export const del = (uuid: uuid) => http.delete(`/archives/${uuid}`)
