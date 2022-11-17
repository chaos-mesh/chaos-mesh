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
import _ from 'lodash'

export const podPhases = ['Pending', 'Running', 'Succeeded', 'Failed', 'Unknown']

export const scopeInitialValues = ({ hasSelector }: { hasSelector: boolean }) => ({
  ...(hasSelector && {
    selector: {
      namespaces: [],
      labelSelectors: [],
      annotationSelectors: [],
      podPhaseSelectors: [],
      pods: [],
      physicalMachines: [],
    },
  }),
  mode: 'all',
  value: undefined,
})

export interface Schedule {
  schedule: string
  historyLimit?: number
  concurrencyPolicy?: 'Forbid' | 'Allow'
  startingDeadlineSeconds?: number
}

export const scheduleInitialValues: Schedule = {
  schedule: '',
  historyLimit: 1,
  concurrencyPolicy: 'Forbid',
  startingDeadlineSeconds: 0,
}

export const removeScheduleValues = (values: any) =>
  _.omit(values, ['scheduled', 'schedule', 'historyLimit', 'concurrencyPolicy', 'startingDeadlineSeconds'])

export const workflowNodeInfoInitialValues = {
  name: '',
  deadline: '',
}
