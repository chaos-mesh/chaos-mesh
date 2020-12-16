export interface EventPod {
  id: string
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
