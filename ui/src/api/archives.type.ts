import { AxiosResponse } from 'axios'

export interface GetArchives {
  (namespace?: string, name?: string, kind?: string): Promise<AxiosResponse<Archive[]>>
}

export interface Archive {
  uid: uuid
  kind: string
  namespace: string
  name: string
  start_time: string
  finish_time: string
}

export type GetArchivesParams = Parameters<GetArchives>

export interface ArchiveDetail extends Archive {
  yaml: any
}
