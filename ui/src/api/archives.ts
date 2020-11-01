import { GetArchiveDetail, GetArchiveReport, GetArchives } from './archives.type'

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

export const detail: GetArchiveDetail = (uuid: uuid) =>
  http.get(`/archives/detail`, {
    params: {
      uid: uuid,
    },
  })

export const report: GetArchiveReport = (uuid: uuid) =>
  http.get(`/archives/report`, {
    params: {
      uid: uuid,
    },
  })
