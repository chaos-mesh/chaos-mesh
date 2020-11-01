/**
 * All variable name of mock data must be <api-name><suffix>, and <suffix> can be ResMock, ParamsMock or RequiredParams
 */
import Mock from 'mockjs'

const Random = Mock.Random

export const archivesResMock = [
  Mock.mock({
    action: '@string',
    archived: '@boolean',
    created_at: '@string',
    deleted_at: '@string',
    finish_time: '@string',
    id: '@natural',
    kind: '@string',
    name: '@string',
    namespace: '@string',
    start_time: '@string',
    uid: '@guid',
    updated_at: '@string',
  }),
]

export const archivesParamsMock = Mock.mock({
  namespace: '@string',
  name: '@string',
  kind: '@string',
})

export const archivesRequiredParams = []

export const archiveDetailResMock = Mock.mock({
  action: '@string',
  finish_time: '@string',
  kind: '@string',
  name: '@string',
  namespace: '@string',
  start_time: '@string',
  uid: '@string',
  yaml: {
    apiVersion: '@string',
    kind: '@string',
    metadata: {
      annotations: {
        additionalProp1: '@string',
        additionalProp2: '@string',
        additionalProp3: '@string',
      },
      labels: {
        additionalProp1: '@string',
        additionalProp2: '@string',
        additionalProp3: '@string',
      },
      name: '@string',
      namespace: '@string',
    },
    spec: {},
  },
})

export const archiveDetailParamsMock = Mock.mock({
  uid: '@guid',
})

export const archiveDetailRequiredParams = ['uid']

export const archiveReportResMock = Mock.mock({
  events: [
    {
      created_at: '@string',
      deleted_at: '@string',
      duration: '@string',
      experiment: '@string',
      experiment_id: '@string',
      finish_time: '@string',
      id: '@natural',
      kind: '@string',
      message: '@string',
      namespace: '@string',
      pods: [
        {
          action: '@string',
          created_at: '@string',
          deleted_at: '@string',
          event_id: '@natural',
          id: '@natural',
          message: '@string',
          namespace: '@string',
          pod_ip: '@string',
          pod_name: '@string',
          updated_at: '@string',
        },
      ],
      start_time: '@string',
      updated_at: '@string',
    },
  ],
  meta: {
    action: '@string',
    archived: '@boolean',
    created_at: '@string',
    deleted_at: '@string',
    finish_time: '@string',
    id: '@natural',
    kind: '@string',
    name: '@string',
    namespace: '@string',
    start_time: '@string',
    uid: '@string',
    updated_at: '@string',
  },
  total_fault_time: '@string',
  total_time: '@string',
})

export const archiveReportParamsMock = Mock.mock({
  uid: '@guid',
})

export const archiveReportRequiredParams = ['uid']

export const eventsResMock = [
  Mock.mock({
    created_at: '@string',
    deleted_at: '@string',
    duration: '@string',
    experiment: '@string',
    experiment_id: '@string',
    finish_time: '@string',
    id: '@natural',
    kind: '@string',
    message: '@string',
    namespace: '@string',
    pods: [
      {
        action: '@string',
        created_at: '@string',
        deleted_at: '@string',
        event_id: '@natural',
        id: '@natural',
        message: '@string',
        namespace: '@string',
        pod_ip: '@string',
        pod_name: '@string',
        updated_at: '@string',
      },
    ],
    start_time: '@string',
    updated_at: '@string',
  }),
]

export const eventsParamsMock = Mock.mock({
  podName: '@string',
  podNamespace: '@string',
  startTime: '@string',
  endTime: '@string',
  experimentName: '@string',
  experimentNamespace: '@string',
  uid: '@string',
  kind: '@string',
  limit: Random.natural().toString(),
})

export const eventsRequiredParams = []

export const experimentsResMock = [
  Mock.mock({
    created: '@string',
    failed_message: '@string',
    kind: '@string',
    name: '@string',
    namespace: '@string',
    status: '@string',
    uid: '@string',
  }),
]

export const experimentsParamsMock = Mock.mock({
  namespace: '@string',
  name: '@string',
  kind: '@string',
  status: '@string',
})

export const experimentsRequiredParams = []
