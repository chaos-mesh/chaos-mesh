/*
 * Copyright 2022 Chaos Mesh Authors.
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
import { array, object, string } from 'yup'

import { Belong } from '.'

const scopeInitialValuesSchema = {
  selector: object({
    namespaces: array().min(1),
  }),
}

export function isInstant(kind: string, action?: string) {
  if (kind === 'PodChaos' && (action === 'pod-kill' || action === 'container-kill')) {
    return true
  }

  return false
}

const workflowNodeInfoSchema = (kind: string, action?: string) => ({
  name: string().trim().required(),
  ...(!isInstant(kind, action) && { deadline: string().trim().required() }),
})

const workflowSchema = (kind: string, action?: string) => {
  return kind !== 'Suspend'
    ? object({
        ...(kind !== 'PhysicalMachineChaos' && scopeInitialValuesSchema),
        ...workflowNodeInfoSchema(kind, action),
      })
    : null
}

export function chooseSchemaByBelong(belong: Belong, kind: string, action?: string) {
  switch (belong) {
    case Belong.Workflow:
      return workflowSchema(kind, action)
  }
}
