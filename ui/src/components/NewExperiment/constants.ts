import * as Yup from 'yup'

import { Experiment } from './types'

export const defaultExperimentSchema: Experiment = {
  name: '',
  namespace: 'default',
  labels: [],
  annotations: [],
  scope: {
    namespace_selectors: ['default'],
    label_selectors: [],
    annotation_selectors: [],
    phase_selectors: ['all'],
    mode: 'one',
    value: '',
    pods: {},
  },
  target: {
    kind: '',
    pod_chaos: {
      action: '',
      container_name: '',
    },
    network_chaos: {
      action: '',
      direction: '',
      bandwidth: {
        buffer: 1,
        limit: 1,
        minburst: 0,
        peakrate: 0,
        rate: '',
      },
      corrupt: {
        correlation: '',
        corrupt: '',
      },
      delay: {
        correlation: '',
        jitter: '',
        latency: '',
      },
      duplicate: {
        correlation: '',
        duplicate: '',
      },
      loss: {
        correlation: '',
        loss: '',
      },
      target: undefined,
    },
    io_chaos: {
      action: '',
      addr: '',
      delay: '',
      errno: '',
      methods: [],
      path: '',
      percent: '100',
    },
    kernel_chaos: {
      fail_kern_request: {
        callchain: [],
        failtype: 0,
        headers: [],
        probability: 0,
        times: 0,
      },
    },
    time_chaos: {
      clock_ids: [],
      container_names: [],
      time_offset: '',
    },
    stress_chaos: {
      stressng_stressors: '',
      stressors: {
        cpu: {
          workers: 1,
          load: 0,
          options: [],
        },
        memory: {
          workers: 1,
          options: [],
        },
      },
      container_name: '',
    },
  },
  scheduler: {
    cron: '',
    duration: '',
  },
}

export const validationSchema = Yup.object({
  name: Yup.string().required('The experiment name is required.'),
  target: Yup.object().when('name', (name: string, schema: Yup.ObjectSchema) =>
    name
      ? Yup.object({
          kind: Yup.string(),
          pod_chaos: Yup.object().when('kind', (kind: string, schema: Yup.ObjectSchema) =>
            kind === 'PodChaos'
              ? Yup.object({
                  action: Yup.string(),
                  container_name: Yup.string().when('action', (action: string, schema: Yup.StringSchema) =>
                    action === 'container-kill' ? schema.required('The Container name is required.') : schema
                  ),
                })
              : schema
          ),
          network_chaos: Yup.object().when('kind', (kind: string, schema: Yup.ObjectSchema) =>
            kind === 'NetworkChaos'
              ? Yup.object({
                  action: Yup.string().required('The NetworkChaos action is required.'),
                  direction: Yup.string().when('action', (action: string, schema: Yup.StringSchema) =>
                    action === 'partition' ? schema.required('The direction is required.') : schema
                  ),
                  bandwidth: Yup.object()
                    .nullable()
                    .when('action', (action: string, schema: Yup.ObjectSchema) =>
                      action === 'bandwidth'
                        ? Yup.object({
                            rate: Yup.string().required('The rate of bandwidth is required.'),
                          })
                        : schema
                    ),
                  corrupt: Yup.object()
                    .nullable()
                    .when('action', (action: string, schema: Yup.ObjectSchema) =>
                      action === 'corrupt'
                        ? Yup.object({
                            corrupt: Yup.string().required('The corrupt is required.'),
                            correlation: Yup.string().required('The correlation of corrupt is required.'),
                          })
                        : schema
                    ),
                  delay: Yup.object()
                    .nullable()
                    .when('action', (action: string, schema: Yup.ObjectSchema) =>
                      action === 'delay'
                        ? Yup.object({
                            latency: Yup.string().required('The latency of delay is required.'),
                          })
                        : schema
                    ),
                  duplicate: Yup.object()
                    .nullable()
                    .when('action', (action: string, schema: Yup.ObjectSchema) =>
                      action === 'duplicate'
                        ? Yup.object({
                            duplicate: Yup.string().required('The duplicate is required.'),
                            correlation: Yup.string().required('The correlation of duplicate is required.'),
                          })
                        : schema
                    ),
                  loss: Yup.object()
                    .nullable()
                    .when('action', (action: string, schema: Yup.ObjectSchema) =>
                      action === 'loss'
                        ? Yup.object({
                            loss: Yup.string().required('The loss is required.'),
                            correlation: Yup.string().required('The correlation of loss is required.'),
                          })
                        : schema
                    ),
                })
              : schema
          ),
          io_chaos: Yup.object().when('kind', (kind: string, schema: Yup.ObjectSchema) =>
            kind === 'IoChaos'
              ? Yup.object({
                  action: Yup.string().required('The IOChaos action is required.'),
                })
              : schema
          ),
          time_chaos: Yup.object().when('kind', (kind: string, schema: Yup.ObjectSchema) =>
            kind === 'TimeChaos'
              ? Yup.object({
                  time_offset: Yup.string().required('The time offset is required.'),
                })
              : schema
          ),
        })
      : schema
  ),
})
