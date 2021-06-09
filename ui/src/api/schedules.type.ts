import { Experiment } from './experiments.type'

export type Schedule = { is: 'schedule' } & Omit<Experiment, 'is' | 'status'>

export interface ScheduleSingle extends Schedule {
  experiment_uids: uuid[]
  kube_object: any
}
