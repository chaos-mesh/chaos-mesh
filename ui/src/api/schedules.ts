/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

import { Schedule, ScheduleParams, ScheduleSingle } from './schedules.type'

import { Archive } from './archives.type'
import { Experiment } from 'components/NewExperiment/types'
import http from './http'

export const newSchedule = (data: Experiment<any>) => http.post('/schedules', data)

export const schedules = (params?: ScheduleParams) =>
  http.get<Schedule[]>('/schedules', {
    params,
  })

export const single = (uuid: uuid) => http.get<ScheduleSingle>(`/schedules/${uuid}`)

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
