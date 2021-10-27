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
import { Experiment as ExperimentResponse, ExperimentSingle, StatusOfExperiments } from './experiments.type'

import { Experiment } from 'components/NewExperiment/types'
import http from './http'

export const state = (namespace = null) =>
  http.get<StatusOfExperiments>('/experiments/state', {
    params: {
      namespace,
    },
  })

export const newExperiment = (data: Experiment<any>) => http.post('/experiments', data)

export const experiments = (namespace = null, name = null, kind = null) =>
  http.get<ExperimentResponse[]>('/experiments', {
    params: {
      namespace,
      name,
      kind,
    },
  })

export const single = (uuid: uuid) => http.get<ExperimentSingle>(`/experiments/${uuid}`)

export const update = (data: ExperimentSingle['kube_object']) => http.put('/experiments', data)

export const del = (uuid: uuid) => http.delete(`/experiments/${uuid}`)
export const delMulti = (uuids: uuid[]) => http.delete(`/experiments?uids=${uuids.join(',')}`)

export const pause = (uuid: uuid) => http.put(`/experiments/pause/${uuid}`)
export const start = (uuid: uuid) => http.put(`/experiments/start/${uuid}`)
