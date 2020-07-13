import { Event } from './events.type'

export interface StateOfExperiments {
  total: number
  running: number
  waiting: number
  paused: number
  failed: number
  finished: number
}

export interface Experiment {
  Kind: string
  Namespace: string
  Name: string
  created: string
  status: keyof StateOfExperiments
  events?: Event[]
}
