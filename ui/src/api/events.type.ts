export interface Event {
  ID: string
  CreateAt: string
  UpdateAt: string
  DeleteAt: string | null
  Experiment: string
  Namespace: string
  Kind: string
  Message: string
  StartTime: string
  FinishTime: string
  Pods: {
    ID: string
    CreateAt: string
    UpdateAt: string
    DeleteAt: string | null
    EventID: number
    PodIP: string
    PodName: string
    Namespace: string
    Message: string
    Action: string
  }[]
}
