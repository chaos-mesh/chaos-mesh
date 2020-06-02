import { Experiment } from 'components/NewExperiment/types'
import http from './http'
import { AxiosResponse } from 'axios'
import { Experiment as ExperimentReponse } from './experiments.type'

export const state = () => http.get('/experiments/state')

export const newExperiment = (data: Experiment) => http.post('/experiments/new', data)

export const experiments: (
  namespace?: string,
  name?: string,
  kind?: string,
  status?: string
) => Promise<AxiosResponse<ExperimentReponse[]>> = (namespace = '', name = '', kind = '', status = '') =>
  http.get(`/experiments?namespace=${namespace}&name=${name}&kind=${kind}&status=${status}`)
