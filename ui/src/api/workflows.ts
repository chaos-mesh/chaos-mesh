import { Workflow, WorkflowDetail, workflowParams } from './workflows.type'

import http from './http'

export const workflows = (params?: workflowParams) =>
  http.get<Workflow[]>('/workflows', {
    params,
  })

export const newWorkflow = (data: any) => http.post('/workflows', data)

export const detail = (ns: string, name: string) => http.get<WorkflowDetail>(`/workflows/${ns}/${name}`)

export const del = (ns: string, name: string) => http.delete(`/workflows/${ns}/${name}`)

export const update = (ns: string, name: string, data: WorkflowDetail['kube_object']) =>
  http.put(`/workflows/${ns}/${name}`, data)
