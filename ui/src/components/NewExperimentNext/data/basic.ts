import * as Yup from 'yup'

const data = {
  name: '',
  namespace: '',
  labels: [],
  annotations: [],
  scope: {
    namespaces: [],
    label_selectors: [],
    annotation_selectors: [],
    phase_selectors: ['all'],
    mode: 'one',
    value: '',
    pods: [],
  },
  scheduler: {
    duration: '',
  },
}

export const schema: Yup.ObjectSchema = Yup.object({
  name: Yup.string().trim().required('The name is required'),
  scope: Yup.object({
    namespaces: Yup.array().min(1, 'The namespace selectors is required'),
  }),
})

export type dataType = typeof data

export default data
