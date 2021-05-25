import { Experiment } from 'components/NewExperiment/types'
import { ScheduleSpecific } from 'components/Schedule/types'
import http from './http'

export const newSchedule = (data: Experiment & ScheduleSpecific) => http.post('/schedules', data)
