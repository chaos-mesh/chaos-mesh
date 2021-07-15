import * as Yup from 'yup'

const data = {
  metadata: {
    name: '',
    namespace: '',
    labels: [],
    annotations: [],
  },
  spec: {
    selector: {
      namespaces: [],
      labelSelectors: [],
      annotationSelectors: [],
      phaseSelectors: ['all'],
      pods: [],
    },
    mode: 'one',
    value: undefined,
    duration: '',
  },
}

export const schema = Yup.object({
  metadata: Yup.object({
    name: Yup.string().trim().required('The name is required'),
  }),
  spec: Yup.object({
    selector: Yup.object({
      namespaces: Yup.array().min(1, 'The namespace selectors is required'),
    }),
  }),
})

export type dataType = typeof data

export default data
