import { ExperimentKind } from 'components/NewExperiment/types'

export interface StatusOfExperiments {
  injecting: number
  running: number
  finished: number
  paused: number
}

export interface Experiment {
  is: 'experiment'
  uid: uuid
  kind: ExperimentKind
  namespace: string
  name: string
  created_at: string
  // FIXME: support keyof in ts-interface-builder
  status: 'injecting' | 'running' | 'finished' | 'paused'
}

export interface ExperimentSingle extends Experiment {
  failed_message: string
  kube_object: any
}
