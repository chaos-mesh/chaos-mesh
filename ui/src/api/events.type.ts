export interface Event {
  id: number
  experiment_id: uuid
  deleted_at: string | null
  experiment: string
  namespace: string
  kind: string
  message: string
  start_time: string
  finish_time: string
  pods:
    | {
        id: string
        delete_at: string | null
        event_id: number
        pod_ip: string
        pod_name: string
        namespace: string
        message: string
        action: string
      }[]
    | null
}
