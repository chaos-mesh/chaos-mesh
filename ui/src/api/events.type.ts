import { ExperimentKind } from '../components/NewExperiment/types'

export interface EventsParams {
  namespace?: string
  limit?: number
}

export interface EventPod {
  id: number
  pod_ip: string
  pod_name: string
  namespace: string
  action: string
  message: string
}

export interface Event {
  id: number
  experiment_id: uuid
  experiment: string
  namespace: string
  kind: ExperimentKind
  message: string
  start_time: string
  finish_time: string
  pods?: EventPod[]
}
