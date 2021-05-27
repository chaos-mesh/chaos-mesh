import { Experiment } from './experiments.type'

// TODO: unify field names
export type Archive = Omit<Experiment, 'status' | 'events'> & {
  start_time: string
  finish_time: string
}

export interface ArchiveDetail extends Archive {
  kube_object: any
}
