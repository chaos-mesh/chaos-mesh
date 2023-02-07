/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import * as Yup from 'yup'

import { Env } from 'slices/experiments'

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
      podPhaseSelectors: [],
      pods: [],
      physicalMachines: [],
    },
    mode: 'all',
    value: undefined,
    address: [],
    duration: '',
  },
}

export const schema = (options: { env: Env; scopeDisabled: boolean; scheduled?: boolean; needDeadline?: boolean }) => {
  let result = Yup.object({
    metadata: Yup.object({
      name: Yup.string().trim().required('The name is required'),
    }),
  })

  const { env, scopeDisabled, scheduled, needDeadline } = options
  let spec = Yup.object()

  if (!scopeDisabled && env === 'k8s') {
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
