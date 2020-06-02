import { Experiment } from './types'

export const defaultExperimentSchema: Experiment = {
  name: '',
  namespace: 'default',
  scope: {
    namespace_selectors: ['default'],
    label_selectors: '{}',
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
        latency: '',
        correlation: '',
        jitter: '',
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
  },
  scheduler: {
    cron: '',
    duration: '',
  },
}
