import { ExperimentDetail, Experiment as ExperimentReponse, StateOfExperiments } from './experiments.type'

import { Experiment } from 'components/NewExperiment/types'
import http from './http'

export const state = (namespace = null) =>
  http.get<StateOfExperiments>('/experiments/state', {
    params: {
      namespace,
    },
  })

export const newExperiment = (data: Experiment) => http.post('/experiments/new', data)

export const experiments = (namespace = null, name = null, kind = null, status = null) =>
  http.get<ExperimentReponse[]>('/experiments', {
    params: {
      namespace,
      name,
      kind,
      status,
    },
  })

export const deleteExperiment = (uuid: uuid) => http.delete(`/experiments/${uuid}`)

export const pauseExperiment = (uuid: uuid) => http.put(`/experiments/pause/${uuid}`)

export const startExperiment = (uuid: uuid) => http.put(`/experiments/start/${uuid}`)

export const detail = (uuid: uuid) => http.get<ExperimentDetail>(`/experiments/detail/${uuid}`)

export const update = (data: ExperimentDetail['yaml']) => http.put('/experiments/update', data)
