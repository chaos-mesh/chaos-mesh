import { Event } from './events.type'

export interface StateOfExperiments {
  Running: number
  Waiting: number
  Paused: number
  Failed: number
  Finished: number
}

export enum StateOfExperimentsEnum {
  Running = 'Running',
  Waiting = 'Waiting',
  Paused = 'Paused',
  Failed = 'Failed',
  Finished = 'Finished',
}

export interface Experiment {
  uid: uuid
  kind: string
  namespace: string
  name: string
  created: string
  status: keyof StateOfExperiments
  events?: Event[]
}

export interface ExperimentDetail extends Experiment {
  failed_message: string
  yaml: any
}
