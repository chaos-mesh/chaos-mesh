import { Experiment } from './experiments.type'

type Common = Omit<Experiment, 'is' | 'status'>

export type Archive = { is: 'archive' } & Common

export interface ArchiveSingle extends Common {
  kube_object: any
}
