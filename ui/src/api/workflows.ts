import { Workflow, WorkflowParams, WorkflowSingle } from './workflows.type'

import { Archive } from './archives.type'
import http from './http'

export const newWorkflow = (data: any) => http.post('/workflows', data)

export const workflows = (params?: WorkflowParams) =>
  http.get<Workflow[]>('/workflows', {
    params,
  })

export const single = (uuid: uuid) => http.get<WorkflowSingle>(`/workflows/${uuid}`)

export const update = (uuid: uuid, data: WorkflowSingle['kube_object']) => http.put(`/workflows/${uuid}`, data)

export const del = (uuid: uuid) => http.delete(`/workflows/${uuid}`)

export const archives = (namespace = null, name = null) =>
  http.get<Archive[]>('/archives/workflows', {
    params: {
      namespace,
      name,
    },
  })

export const singleArchive = (uuid: uuid) => http.get<Archive>(`archives/workflows/${uuid}`)

export const delArchive = (uuid: uuid) => http.delete(`/archives/workflows/${uuid}`)
export const delArchives = (uuids: uuid[]) => http.delete(`/archives/workflows?uids=${uuids.join(',')}`)
