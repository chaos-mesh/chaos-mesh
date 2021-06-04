import { ExperimentKind } from '../components/NewExperiment/types'

export interface EventsParams {
  object_id?: uuid
  namespace?: string
  limit?: number
}

export interface Event {
  id: number
  object_id: uuid
  namespace: string
  name: string
  kind: ExperimentKind | 'Schedule'
  type: 'Normal' | 'Warning'
  reason: string
  created_at: string
  message: string
}
