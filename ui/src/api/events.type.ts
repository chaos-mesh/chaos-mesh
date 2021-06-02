import { ExperimentKind } from '../components/NewExperiment/types'

export interface EventsParams {
  uid?: uuid
  namespace?: string
  limit?: number
}

export interface Event {
  id: number
  object_id: uuid
  name: string
  namespace: string
  kind: ExperimentKind | 'Schedule'
  type: 'Normal' | 'Warning'
  created_at: string
  message: string
}
