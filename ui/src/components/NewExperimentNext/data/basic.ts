import * as Yup from 'yup'

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

export const schema: Record<string, Yup.Schema<any>> = {
  name: Yup.string().required('The experiment name is required'),
}

export type dataType = typeof data

export default data
