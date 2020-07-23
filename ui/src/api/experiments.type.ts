import { Event } from './events.type'

export interface StateOfExperiments {
  Total: number
  Running: number
  Waiting: number
  Paused: number
  Failed: number
  Finished: number
}

export enum StateOfExperimentsEnum {
  Total = 'total',
  Running = 'running',
  Waiting = 'waiting',
  Paused = 'paused',
  Failed = 'failed',
  Finished = 'finished',
}

export interface Experiment {
  Kind: string
  Namespace: string
  Name: string
  created: string
  status: keyof StateOfExperiments
  uid: uuid
  events?: Event[]
}
