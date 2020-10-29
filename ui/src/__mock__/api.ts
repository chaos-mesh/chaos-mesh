import Mock from 'mockjs'
import { Pool } from '@material-ui/icons'

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

export const archivesDetailResMock = Mock.mock({
  action: '@string',
  archived: '@boolean',
  created_at: '@string',
  deleted_at: '@string',
  experiment_info: {
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
    scheduler: {
      cron: '@string',
      duration: '@string',
    },
    scope: {
      annotation_selectors: {
        additionalProp1: '@string',
        additionalProp2: '@string',
        additionalProp3: '@string',
      },
      field_selectors: {
        additionalProp1: '@string',
        additionalProp2: '@string',
        additionalProp3: '@string',
      },
      label_selectors: {
        additionalProp1: '@string',
        additionalProp2: '@string',
        additionalProp3: '@string',
      },
      mode: '@string',
      namespace_selectors: ['@string'],
      phase_selectors: ['@string'],
      pods: {
        additionalProp1: ['@string'],
        additionalProp2: ['@string'],
        additionalProp3: ['@string'],
      },
      value: '@string',
    },
    target: {
      io_chaos: {
        action: '@string',
        addr: '@string',
        delay: '@string',
        errno: '@string',
        methods: ['@string'],
        path: '@string',
        percent: '@string',
      },
      kernel_chaos: {
        fail_kern_request: {
          callchain: [
            {
              funcname: '@string',
              parameters: '@string',
              predicate: '@string',
            },
          ],
          failtype: '@natural',
          headers: ['@string'],
          probability: '@natural',
          times: '@natural',
        },
      },
      kind: '@string',
      network_chaos: {
        action: '@string',
        bandwidth: {
          buffer: '@natural',
          limit: '@natural',
          minburst: '@natural',
          peakrate: '@natural',
          rate: '@string',
        },
        corrupt: {
          correlation: '@string',
          corrupt: '@string',
        },
        delay: {
          correlation: '@string',
          jitter: '@string',
          latency: '@string',
          reorder: {
            correlation: '@string',
            gap: '@natural',
            reorder: '@string',
          },
        },
        direction: '@string',
        duplicate: {
          correlation: '@string',
          duplicate: '@string',
        },
        loss: {
          correlation: '@string',
          loss: '@string',
        },
        target_scope: {
          annotation_selectors: {
            additionalProp1: '@string',
            additionalProp2: '@string',
            additionalProp3: '@string',
          },
          field_selectors: {
            additionalProp1: '@string',
            additionalProp2: '@string',
            additionalProp3: '@string',
          },
          label_selectors: {
            additionalProp1: '@string',
            additionalProp2: '@string',
            additionalProp3: '@string',
          },
          mode: '@string',
          namespace_selectors: ['@string'],
          phase_selectors: ['@string'],
          pods: {
            additionalProp1: ['@string'],
            additionalProp2: ['@string'],
            additionalProp3: ['@string'],
          },
          value: '@string',
        },
      },
      pod_chaos: {
        action: '@string',
        container_name: '@string',
      },
      stress_chaos: {
        container_name: '@string',
        stressng_stressors: '@string',
        stressors: {
          cpu: {
            load: '@natural',
            options: ['@string'],
            workers: '@natural',
          },
          memory: {
            options: ['@string'],
            workers: '@natural',
          },
        },
      },
      time_chaos: {
        clock_ids: ['@string'],
        container_names: ['@string'],
        time_offset: '@string',
      },
    },
  },
  finish_time: '@string',
  id: '@natural',
  kind: '@string',
  name: '@string',
  namespace: '@string',
  start_time: '@string',
  uid: '@string',
  updated_at: '@string',
})

export const archivesDetailParamsMock = Mock.mock({
  uid: '@guid',
})

export const archivesDetailRequiredParams = ['uid']

export const archivesReportResMock = Mock.mock({
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

export const archivesReportParamsMock = Mock.mock({
  uid: '@guid',
})

export const archivesReportRequiredParams = ['uid']

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

export const dryEventsResMock = []
