import { Archive } from './archives.type'
import { Experiment } from 'components/NewExperiment/types'
import { Schedule } from './schedules.type'
import { ScheduleSpecific } from 'components/Schedule/types'
import http from './http'

export const newSchedule = (data: Experiment & ScheduleSpecific) => http.post('/schedules', data)

export const schedules = (namespace = null) =>
  http.get<Schedule[]>('/schedules', {
    params: {
      namespace,
    },
  })

export const del = (uuid: uuid) => http.delete(`/schedules/${uuid}`)

export const archives = (namespace = null, name = null, kind = null) =>
  http.get<Archive[]>('/archives/schedules', {
    params: {
      namespace,
      name,
      kind,
    },
  })
