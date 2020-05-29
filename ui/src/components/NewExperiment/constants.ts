import { Experiment } from './types'

export const defaultExperimentSchema: Experiment = {
  basic: {
    name: '',
    namespace: '',
  },
  scope: {
    namespaceSelector: [],
    phaseSelector: [],
    mode: 'all',
    value: '',
  },
  target: {
    pod: {
      action: '',
      container: '',
    },
    network: {
      action: '',
      delay: {
        latency: '',
        correlation: '',
        jitter: '',
      },
    },
  },
  schedule: {
    cron: '',
    duration: '',
  },
}
