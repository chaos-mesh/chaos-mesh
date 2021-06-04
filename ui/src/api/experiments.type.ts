import { Event } from './events.type'
import { ExperimentKind } from 'components/NewExperiment/types'

export interface StateOfExperiments {
  injecting: number
  running: number
  finished: number
  paused: number
}

export interface Experiment {
  uid: uuid
  kind: ExperimentKind
  namespace: string
  name: string
  created_at: string
  // FIXME: support keyof in ts-interface-builder
  status: 'injecting' | 'running' | 'finished' | 'paused'
  events?: Event[]
}

export interface ExperimentSingle extends Experiment {
  failed_message: string
  kube_object: any
}
