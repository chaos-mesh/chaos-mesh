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
