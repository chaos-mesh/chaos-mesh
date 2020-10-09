export interface Archive {
  kind: string
  namespace: string
  name: string
  uid: uuid
  start_time: string
  finish_time: string
}

export interface ArchiveDetail extends Archive {
  yaml: any
}
