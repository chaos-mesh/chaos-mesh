import { AxiosResponse } from 'axios'

export interface GetArchives {
  (namespace?: string, name?: string, kind?: string): Promise<AxiosResponse<Archive[]>>
}
/**
 * @description archives:result
 */
export interface Archive {
  uid: uuid
  kind: string
  namespace: string
  name: string
  start_time: string
  finish_time: string
}

/**
 * @description archives:params
 */
export type GetArchivesParams = Parameters<GetArchives>

export interface GetArchiveDetail {
  (uid: string): Promise<AxiosResponse<ArchiveDetail>>
}

/**
 * @description archiveDetail:result
 */
export interface ArchiveDetail extends Archive {
  yaml: any
}
/**
 * @description archiveDetail:params
 */
export type GetArchiveDetailParams = Parameters<GetArchiveDetail>

export interface GetArchiveReport {
  (uid: string): Promise<AxiosResponse<ArchiveReport>>
}

/**
 * @description archiveReport:result
 */
export interface ArchiveReport {
  events: any
  meta: Archive
}
/**
 * @description archiveReport:params
 */
export type GetArchiveReportParams = Parameters<GetArchiveReport>
