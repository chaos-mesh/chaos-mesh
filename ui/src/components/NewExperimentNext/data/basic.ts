import * as Yup from 'yup'

import { schema as scheduleSchema } from 'components/Schedule/types'

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

export const schema = (options: { scopeDisabled: boolean; scheduled?: boolean; needDeadline?: boolean }) => {
  let result = Yup.object({
    metadata: Yup.object({
      name: Yup.string().trim().required('The name is required'),
    }),
  })

  const { scopeDisabled, scheduled, needDeadline } = options
  let spec = Yup.object()

  if (!scopeDisabled) {
    spec = spec.shape({
      selector: Yup.object({
        namespaces: Yup.array().min(1, 'The namespace selectors is required'),
      }),
    })
  }

  if (scheduled) {
    spec = spec.shape(scheduleSchema)
  }

  if (needDeadline) {
    spec = spec.shape({
      duration: Yup.string().trim().required('The deadline is required'),
    })
  }

  return result.shape({
    spec,
  })
}

export type dataType = typeof data

export default data
