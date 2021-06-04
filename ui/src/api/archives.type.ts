import { Experiment } from './experiments.type'

// TODO: unify field names
export type Archive = Omit<Experiment, 'status' | 'events'>

export interface ArchiveDetail extends Archive {
  kube_object: any
}
