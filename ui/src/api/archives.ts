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

export const single = (uuid: uuid) =>
  http.get('/archives/detail', {
    params: {
      uid: uuid,
    },
  })

export const del = (uuid: uuid) => http.delete(`/archives/${uuid}`)
export const delMulti = (uuids: uuid[]) => http.delete(`/archives?uids=${uuids.join(',')}`)
