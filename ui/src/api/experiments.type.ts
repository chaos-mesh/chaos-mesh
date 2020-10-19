import { Event } from './events.type'
import { Experiment as ExperimentInfo } from 'components/NewExperiment/types'

export interface StateOfExperiments {
  Total: number
  Running: number
  Waiting: number
  Paused: number
  Failed: number
  Finished: number
}

export enum StateOfExperimentsEnum {
  Total = 'Total',
  Running = 'Running',
  Waiting = 'Waiting',
  Paused = 'Paused',
  Failed = 'Failed',
  Finished = 'Finished',
}

export interface Experiment {
  kind: string
  namespace: string
  name: string
  created: string
  status: keyof StateOfExperiments
  uid: uuid
  events?: Event[]
}

export interface ExperimentDetail extends Experiment {
  failed_message: string
  experiment_info: ExperimentInfo
}
