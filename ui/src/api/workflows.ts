import { Workflow, WorkflowDetail, workflowParams } from './workflows.type'

import http from './http'

export const workflows = (params?: workflowParams) =>
  http.get<Workflow[]>('/workflows', {
    params,
  })

export const detail = (ns: string, name: string) => http.get<WorkflowDetail>(`/workflows/detail/${ns}/${name}`)

export const del = (ns: string, name: string) => http.delete(`/workflows/${ns}/${name}`)
