/**
 * All variable name of mock data must be <api-name><suffix>, and <suffix> can be ResMock, ParamsMock or RequiredParams
 */
import Chance from 'chance'

const chance = new Chance()

export const archivesResMock = [
  {
    action: chance.string(),
    archived: chance.bool(),
    created_at: chance.string(),
    deleted_at: chance.string(),
    finish_time: chance.string(),
    id: chance.natural(),
    kind: chance.string(),
    name: chance.string(),
    namespace: chance.string(),
    start_time: chance.string(),
    uid: chance.guid(),
    updated_at: chance.string(),
  },
]

export const archivesParamsMock = {
  namespace: chance.string(),
  name: chance.string(),
  kind: chance.string(),
}

export const archivesRequiredParams = []

export const archiveDetailResMock = {
  action: chance.string(),
  finish_time: chance.string(),
  kind: chance.string(),
  name: chance.string(),
  namespace: chance.string(),
  start_time: chance.string(),
  uid: chance.string(),
  yaml: {
    apiVersion: chance.string(),
    kind: chance.string(),
    metadata: {
      annotations: {
        additionalProp1: chance.string(),
        additionalProp2: chance.string(),
        additionalProp3: chance.string(),
      },
      labels: {
        additionalProp1: chance.string(),
        additionalProp2: chance.string(),
        additionalProp3: chance.string(),
      },
      name: chance.string(),
      namespace: chance.string(),
    },
    spec: {},
  },
}

export const archiveDetailParamsMock = {
  uid: chance.guid(),
}

export const archiveDetailRequiredParams = ['uid']

export const archiveReportResMock = {
  events: [
    {
      created_at: chance.string(),
      deleted_at: chance.string(),
      duration: chance.string(),
      experiment: chance.string(),
      experiment_id: chance.string(),
      finish_time: chance.string(),
      id: chance.natural(),
      kind: chance.string(),
      message: chance.string(),
      namespace: chance.string(),
      pods: [
        {
          action: chance.string(),
          created_at: chance.string(),
          deleted_at: chance.string(),
          event_id: chance.natural(),
          id: chance.natural(),
          message: chance.string(),
          namespace: chance.string(),
          pod_ip: chance.string(),
          pod_name: chance.string(),
          updated_at: chance.string(),
        },
      ],
      start_time: chance.string(),
      updated_at: chance.string(),
    },
  ],
  meta: {
    action: chance.string(),
    archived: chance.bool(),
    created_at: chance.string(),
    deleted_at: chance.string(),
    finish_time: chance.string(),
    id: chance.natural(),
    kind: chance.string(),
    name: chance.string(),
    namespace: chance.string(),
    start_time: chance.string(),
    uid: chance.string(),
    updated_at: chance.string(),
  },
  total_fault_time: chance.string(),
  total_time: chance.string(),
}

export const archiveReportParamsMock = {
  uid: chance.guid(),
}

export const archiveReportRequiredParams = ['uid']

export const eventsResMock = [
  {
    created_at: chance.string(),
    deleted_at: chance.string(),
    duration: chance.string(),
    experiment: chance.string(),
    experiment_id: chance.string(),
    finish_time: chance.string(),
    id: chance.natural(),
    kind: chance.string(),
    message: chance.string(),
    namespace: chance.string(),
    pods: [
      {
        action: chance.string(),
        created_at: chance.string(),
        deleted_at: chance.string(),
        event_id: chance.natural(),
        id: chance.natural(),
        message: chance.string(),
        namespace: chance.string(),
        pod_ip: chance.string(),
        pod_name: chance.string(),
        updated_at: chance.string(),
      },
    ],
    start_time: chance.string(),
    updated_at: chance.string(),
  },
]

export const eventsParamsMock = {
  podName: chance.string(),
  podNamespace: chance.string(),
  startTime: chance.string(),
  endTime: chance.string(),
  experimentName: chance.string(),
  experimentNamespace: chance.string(),
  uid: chance.string(),
  kind: chance.string(),
  limit: chance.natural().toString(),
}

export const eventsRequiredParams = []

export const experimentsResMock = [
  {
    created: chance.string(),
    failed_message: chance.string(),
    kind: chance.string(),
    name: chance.string(),
    namespace: chance.string(),
    status: chance.string(),
    uid: chance.string(),
  },
]

export const experimentsParamsMock = {
  namespace: chance.string(),
  name: chance.string(),
  kind: chance.string(),
  status: chance.string(),
}

export const experimentsRequiredParams = []
