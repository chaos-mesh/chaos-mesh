import { ExperimentDetail, Experiment as ExperimentReponse, StateOfExperiments } from './experiments.type'

import { AxiosResponse } from 'axios'
import { Experiment } from 'components/NewExperiment/types'
import http from './http'

export const state: () => Promise<AxiosResponse<StateOfExperiments>> = () => http.get('/experiments/state')

export const newExperiment = (data: Experiment) => http.post('/experiments/new', data)

export const experiments: (
  namespace?: string,
  name?: string,
  kind?: string,
  status?: string
) => Promise<AxiosResponse<ExperimentReponse[]>> = (namespace = '', name = '', kind = '', status = '') =>
  http.get(`/experiments?namespace=${namespace}&name=${name}&kind=${kind}&status=${status}`)

export const deleteExperiment = (uuid: uuid) => http.delete(`/experiments/${uuid}`)

export const pauseExperiment = (uuid: uuid) => http.put(`/experiments/pause/${uuid}`)

export const startExperiment = (uuid: uuid) => http.put(`/experiments/start/${uuid}`)

export const detail: (uuid: uuid) => Promise<AxiosResponse<ExperimentDetail>> = (uuid) =>
  http.get(`/experiments/detail/${uuid}`)

export const update = (data: ExperimentDetail['yaml']) => http.put('/experiments/update', data)
