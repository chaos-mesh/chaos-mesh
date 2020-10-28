const data = {
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
    pods: [],
  },
  scheduler: {
    cron: '',
    duration: '',
  },
}

export type dataType = typeof data

export default data
