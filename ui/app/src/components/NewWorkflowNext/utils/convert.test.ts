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
import { ExperimentKind, nodeExperimentToTemplate, templateTypeToFieldName } from './convert'

describe('components/NewWorkflowNext/utils/convert', () => {
  describe('templateTypeToFieldName', () => {
    it('should return the correct field names', () => {
      expect(templateTypeToFieldName(ExperimentKind.AWSChaos)).toBe('awsChaos')
      expect(templateTypeToFieldName(ExperimentKind.PhysicalMachineChaos)).toBe('physicalmachineChaos')
    })
  })

  describe('nodeExperimentToTemplate', () => {
    it('should return the correct PodChaos template', () => {
      const data = {
        name: 'p1',
        templateType: 'PodChaos',
        deadline: '1m',
        action: 'pod-failure',
        selector: {
          namespaces: ['default'],
        },
        mode: 'all',
      }

      expect(nodeExperimentToTemplate(data)).toEqual({
        name: 'p1',
        templateType: 'PodChaos',
        deadline: '1m',
        podChaos: {
          action: 'pod-failure',
          selector: {
            namespaces: ['default'],
          },
          mode: 'all',
        },
      })
    })

    it('should return the correct Schedule template', () => {
      const data = {
        name: 's1',
        templateType: 'PodChaos',
        deadline: '1m',
        scheduled: true,
        schedule: '@every 2h',
        historyLimit: 1,
        concurrencyPolicy: 'Forbid',
        action: 'pod-failure',
        selector: {
          namespaces: ['default'],
        },
        mode: 'all',
      }

      expect(nodeExperimentToTemplate(data)).toEqual({
        name: 's1',
        templateType: 'Schedule',
        deadline: '1m',
        schedule: {
          schedule: '@every 2h',
          historyLimit: 1,
          concurrencyPolicy: 'Forbid',
          type: 'PodChaos',
          podChaos: {
            action: 'pod-failure',
            selector: {
              namespaces: ['default'],
            },
            mode: 'all',
          },
        },
      })
    })

    it('should return the correct Suspend template', () => {
      const data = {
        name: 's2',
        templateType: 'Suspend',
        deadline: '1m',
      }

      expect(nodeExperimentToTemplate(data)).toEqual({
        name: 's2',
        templateType: 'Suspend',
        deadline: '1m',
      })
    })
  })
})
