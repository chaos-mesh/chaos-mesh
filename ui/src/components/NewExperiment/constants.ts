import { Experiment } from './types'
import * as Yup from 'yup'

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
  },
  target: {
    kind: 'PodChaos',
    pod_chaos: {
      action: '',
      container_name: '',
    },
    network_chaos: {
      action: '',
      bandwidth: {
        buffer: 0,
        limit: 0,
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
      fail_kernel_req: {
        callchain: [],
        failtype: 0,
        headers: [],
        probability: 0,
        times: 0,
      },
    },
  },
  scheduler: {
    cron: '',
    duration: '',
  },
}

export const validationSchema = Yup.object().shape({
  name: Yup.string().required(),
})
