import { Experiment as ExperimentResponse, ExperimentSingle, StatusOfExperiments } from './experiments.type'

import { Experiment } from 'components/NewExperiment/types'
import http from './http'

export const state = (namespace = null) =>
  http.get<StatusOfExperiments>('/experiments/state', {
    params: {
      namespace,
    },
  })

export const newExperiment = (data: Experiment) => http.post('/experiments/new', data)

export const experiments = (namespace = null, name = null, kind = null) =>
  http.get<ExperimentResponse[]>('/experiments', {
    params: {
      namespace,
      name,
      kind,
    },
  })

export const single = (uuid: uuid) => http.get<ExperimentSingle>(`/experiments/detail/${uuid}`)

export const update = (data: ExperimentSingle['kube_object']) => http.put('/experiments/update', data)

export const del = (uuid: uuid) => http.delete(`/experiments/${uuid}`)
export const delMulti = (uuids: uuid[]) => http.delete(`/experiments?uids=${uuids.join(',')}`)

export const pause = (uuid: uuid) => http.put(`/experiments/pause/${uuid}`)
export const start = (uuid: uuid) => http.put(`/experiments/start/${uuid}`)
