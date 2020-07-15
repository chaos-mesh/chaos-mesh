import { Event } from './events.type'

export interface StateOfExperiments {
  total: number
  running: number
  waiting: number
  paused: number
  failed: number
  finished: number
}

export enum StateOfExperimentsEnum {
  total = 'total',
  running = 'running',
  waiting = 'waiting',
  paused = 'paused',
  failed = 'failed',
  finished = 'finished',
}

export interface Experiment {
  Kind: string
  Namespace: string
  Name: string
  created: string
  status: keyof StateOfExperiments
  events?: Event[]
}
