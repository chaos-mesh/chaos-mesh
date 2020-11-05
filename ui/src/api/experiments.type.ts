import { AxiosResponse } from 'axios'
import { Event } from './events.type'

export interface StateOfExperiments {
  Total: number
  Running: number
  Waiting: number
  Paused: number
  Failed: number
  Finished: number
}

export enum StateOfExperimentsEnum {
  Total = 'Total',
  Running = 'Running',
  Waiting = 'Waiting',
  Paused = 'Paused',
  Failed = 'Failed',
  Finished = 'Finished',
}
/**
 * @description experiments:result
 */
export interface ExperimentFromAPIData {
  uid: uuid
  kind: string
  namespace: string
  name: string
  created: string
  status: keyof StateOfExperiments
}

export interface Experiment extends ExperimentFromAPIData {
  events?: Event[]
}

export interface GetExperiment {
  (namespace?: string, name?: string, kind?: string, status?: string): Promise<AxiosResponse<Experiment[]>>
}
/**
 * @description experiments:params
 */
export type GetExperimentParams = Parameters<GetExperiment>

export interface ExperimentDetail extends Experiment {
  failed_message: string
  yaml: any
}
