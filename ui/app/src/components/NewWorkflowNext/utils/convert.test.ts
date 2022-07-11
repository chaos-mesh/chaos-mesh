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
import type { Node } from 'react-flow-renderer'
import { v4 as uuidv4 } from 'uuid'

import {
  ExperimentKind,
  Template,
  connectNodes,
  nodeExperimentToTemplate,
  templateToNodeExperiment,
  templateTypeToFieldName,
  workflowToFlow,
} from './convert'

const nodeExperimentPodChaosSample = {
  kind: 'PodChaos',
  name: 'p1',
  templateType: 'PodChaos',
  deadline: '1m',
  action: 'pod-failure',
  selector: {
    namespaces: ['default'],
  },
  mode: 'all',
}

const templatePodChaosSample: any = {
  name: 'p1',
  templateType: ExperimentKind.PodChaos,
  deadline: '1m',
  podChaos: {
    action: 'pod-failure',
    selector: {
      namespaces: ['default'],
    },
    mode: 'all',
  },
}

const nodeExperimentScheduleSample = {
  kind: 'PodChaos',
  name: 's1',
  templateType: 'PodChaos',
  deadline: '1m',
  scheduled: true,
  schedule: '@every 2h',
  historyLimit: 2,
  concurrencyPolicy: 'Forbid',
  action: 'pod-failure',
  selector: {
    namespaces: ['default'],
  },
  mode: 'all',
}

const templateScheduleSample: Template = {
  name: 's1',
  templateType: 'Schedule',
  deadline: '1m',
  schedule: {
    schedule: '@every 2h',
    historyLimit: 2,
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
}

const workflowSample1 = `
apiVersion: chaos-mesh.org/v1alpha1
kind: Workflow
metadata:
  name: try-workflow-parallel
spec:
  entry: the-entry
  templates:
    - name: the-entry
      templateType: Parallel
      deadline: 240s
      children:
        - workflow-stress-chaos
        - workflow-network-chaos
        - workflow-pod-chaos-schedule
    - name: workflow-network-chaos
      templateType: NetworkChaos
      deadline: 20s
      networkChaos:
        direction: to
        action: delay
        mode: all
        selector:
          labelSelectors:
            'app': 'hello-kubernetes'
        delay:
          latency: '90ms'
          correlation: '25'
          jitter: '90ms'
    - name: workflow-pod-chaos-schedule
      templateType: Schedule
      deadline: 40s
      schedule:
        schedule: '@every 2s'
        type: 'PodChaos'
        podChaos:
          action: pod-kill
          mode: one
          selector:
            labelSelectors:
              'app': 'hello-kubernetes'
    - name: workflow-stress-chaos
      templateType: StressChaos
      deadline: 20s
      stressChaos:
        mode: one
        selector:
          labelSelectors:
            'app': 'hello-kubernetes'
        stressors:
          cpu:
            workers: 1
            load: 20
            options: ['--cpu 1', '--timeout 600']
`

describe('components/NewWorkflowNext/utils/convert', () => {
  describe('templateTypeToFieldName', () => {
    it('should return the correct field names', () => {
      expect(templateTypeToFieldName(ExperimentKind.AWSChaos)).toBe('awsChaos')
      expect(templateTypeToFieldName(ExperimentKind.HTTPChaos)).toBe('httpChaos')
      expect(templateTypeToFieldName(ExperimentKind.PhysicalMachineChaos)).toBe('physicalmachineChaos')
    })
  })

  describe('nodeExperimentToTemplate', () => {
    it('should return the correct PodChaos template', () => {
      expect(nodeExperimentToTemplate(nodeExperimentPodChaosSample)).toEqual(templatePodChaosSample)
    })

    it('should return the correct Schedule template', () => {
      expect(nodeExperimentToTemplate(nodeExperimentScheduleSample)).toEqual(templateScheduleSample)
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

  describe('templateToNodeExperiment', () => {
    it('should return the correct PodChaos NodeExperiment', () => {
      const { id, ...rest } = templateToNodeExperiment(templatePodChaosSample)

      expect(rest).toEqual(nodeExperimentPodChaosSample)
    })

    it('should return the correct Schedule NodeExperiment', () => {
      const { id, ...rest } = templateToNodeExperiment(templateScheduleSample, true)

      expect(rest).toEqual({
        ...nodeExperimentScheduleSample,
        startingDeadlineSeconds: 0,
      })
    })
  })

  describe('connectNodes', () => {
    const nodes: Node[] = [
      {
        id: uuidv4(),
        position: { x: 0, y: 0 },
        data: {},
      },
      {
        id: uuidv4(),
        position: { x: 0, y: 0 },
        data: {},
      },
      {
        id: uuidv4(),
        position: { x: 0, y: 0 },
        data: {},
      },
    ]
    it('should return the correct connections', () => {
      const result = connectNodes(nodes)

      expect(result.length).toBe(2)
      expect(result[0].source).toBe(nodes[0].id)
      expect(result[0].target).toBe(nodes[1].id)
      expect(result[1].source).toBe(nodes[1].id)
      expect(result[1].target).toBe(nodes[2].id)
    })
  })

  describe('workflowToFlow', () => {
    it('test workflow sample 1', () => {
      const { nodes, edges } = workflowToFlow(workflowSample1)

      expect(nodes.length).toBe(4)
      expect(edges.length).toBe(0)
    })
  })
})
