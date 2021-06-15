import { Schedule, ScheduleParams, ScheduleSingle } from './schedules.type'

import { Archive } from './archives.type'
import { Experiment } from 'components/NewExperiment/types'
import { ScheduleSpecific } from 'components/Schedule/types'
import http from './http'

export const newSchedule = (data: Experiment & ScheduleSpecific) => http.post('/schedules', data)

export const schedules = (params?: ScheduleParams) =>
  http.get<Schedule[]>('/schedules', {
    params,
  })

export const single = (uuid: uuid) => http.get<ScheduleSingle>(`/schedules/${uuid}`)

export const update = (data: ScheduleSingle['kube_object']) => http.put('/schedules', data)

export const del = (uuid: uuid) => http.delete(`/schedules/${uuid}`)
export const delMulti = (uuids: uuid[]) => http.delete(`/schedules?uids=${uuids.join(',')}`)

export const pause = (uuid: uuid) => http.put(`/schedules/pause/${uuid}`)
export const start = (uuid: uuid) => http.put(`/schedules/start/${uuid}`)

export const archives = (namespace = null, name = null, kind = null) =>
  http.get<Archive[]>('/archives/schedules', {
    params: {
      namespace,
      name,
      kind,
    },
  })

export const singleArchive = (uuid: uuid) => http.get<Archive>(`archives/schedules/${uuid}`)

export const delArchive = (uuid: uuid) => http.delete(`/archives/schedules/${uuid}`)
export const delArchives = (uuids: uuid[]) => http.delete(`/archives/schedules?uids=${uuids.join(',')}`)
