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

export const schema = (options: { scopeDisabled: boolean }) => {
  let result = Yup.object({
    name: Yup.string().trim().required('The name is required'),
  })

  if (!options.scopeDisabled) {
    result = result.shape({
      scope: Yup.object({
        namespaces: Yup.array().min(1, 'The namespace selectors is required'),
      }),
    })
  }

  return result
}

export type dataType = typeof data

export default data
