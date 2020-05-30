import { Experiment } from 'components/NewExperiment/types'
import http from './http'

export const state = () => http.get('/experiments/state')

export const newExperiment = (data: Experiment) => http.post('/experiments/new', data)
