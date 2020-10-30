import { AxiosResponse } from 'axios'

export interface EventsParams {
  limit?: string
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
  kind: string
  message: string
  start_time: string
  finish_time: string
  pods: EventPod[] | null
}

export interface GetEvents {
  (params?: EventsParams): Promise<AxiosResponse<Event[]>>
}

export type GetEventsParams = Parameters<GetEvents>
