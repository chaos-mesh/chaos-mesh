import { Experiment } from './experiments.type'

export interface ScheduleParams {
  namespace?: string
}

export type Schedule = { is: 'schedule' } & Omit<Experiment, 'is'>

export interface ScheduleSingle extends Schedule {
  experiment_uids: uuid[]
  kube_object: any
}
