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
import yaml from 'js-yaml'
import _ from 'lodash'
import type { Edge } from 'react-flow-renderer'
import { v4 as uuidv4 } from 'uuid'

import { NodeExperiment } from 'slices/workflows'

import { scheduleInitialValues } from 'components/AutoForm/data'

import { isDeepEmpty } from 'lib/utils'

export enum ExperimentKind {
  AWSChaos = 'AWSChaos',
  AzureChaos = 'AzureChaos',
  BlockChaos = 'BlockChaos',
  DNSChaos = 'DNSChaos',
  GCPChaos = 'GCPChaos',
  HTTPChaos = 'HTTPChaos',
  IOChaos = 'IOChaos',
  JVMChaos = 'JVMChaos',
  KernelChaos = 'KernelChaos',
  NetworkChaos = 'NetworkChaos',
  PodChaos = 'PodChaos',
  StressChaos = 'StressChaos',
  TimeChaos = 'TimeChaos',
  PhysicalMachineChaos = 'PhysicalMachineChaos',
}

const mapping = new Map<ExperimentKind, string>([
  [ExperimentKind.AWSChaos, 'awsChaos'],
  [ExperimentKind.AzureChaos, 'azureChaos'],
  [ExperimentKind.BlockChaos, 'blockChaos'],
  [ExperimentKind.DNSChaos, 'dnsChaos'],
  [ExperimentKind.GCPChaos, 'gcpChaos'],
  [ExperimentKind.HTTPChaos, 'httpChaos'],
  [ExperimentKind.IOChaos, 'ioChaos'],
  [ExperimentKind.JVMChaos, 'jvmChaos'],
  [ExperimentKind.KernelChaos, 'kernelChaos'],
  [ExperimentKind.NetworkChaos, 'networkChaos'],
  [ExperimentKind.PodChaos, 'podChaos'],
  [ExperimentKind.StressChaos, 'stressChaos'],
  [ExperimentKind.TimeChaos, 'timeChaos'],
  [ExperimentKind.PhysicalMachineChaos, 'physicalmachineChaos'],
])

export function templateTypeToFieldName(templateType: ExperimentKind): string {
  return mapping.get(templateType)!
}

export enum SpecialTemplateType {
  Serial = 'Serial',
  Parallel = 'Parallel',
  Suspend = 'Suspend',
}

export interface Template {
  id?: uuid
  level?: number
  name: string
  templateType: SpecialTemplateType | ExperimentKind | 'Schedule'
  deadline?: string
  schedule?: { type: string } & typeof scheduleInitialValues
  children?: string[]
}

/**
 * Convert edges to ES6 Map with source node UUID as key and edges array as value.
 *
 * @param {Edge[]} edges
 * @return {Map<uuid, Edge[]>}
 */
function edgesToSourceMap(edges: Edge[]): Map<uuid, Edge[]> {
  const map = new Map()

  edges.forEach((edge) => {
    if (map.has(edge.source)) {
      map.set(edge.source, [...map.get(edge.source), edge])
    } else {
      map.set(edge.source, [edge])
    }
  })

  return map
}

function findNextNodeArray(origin: string, result: uuid[], edgesMap: Map<uuid, Edge[]>): uuid[] {
  if (edgesMap.has(origin)) {
    const target = edgesMap.get(origin)![0].target

    return findNextNodeArray(target, [...result, target], edgesMap)
  }

  return result
}

export function nodeExperimentToTemplate(node: NodeExperiment): Template {
  const { id, kind, name, templateType, deadline, scheduled, ...rest } = JSON.parse(JSON.stringify(node))

  if (scheduled) {
    const { schedule, historyLimit, concurrencyPolicy, startingDeadlineSeconds, ...restrest } = rest

    return {
      name,
      templateType: 'Schedule',
      deadline,
      schedule: {
        schedule,
        historyLimit,
        concurrencyPolicy,
        startingDeadlineSeconds,
        type: templateType,
        [templateTypeToFieldName(templateType)]: restrest,
      },
    }
  }

  const fieldName = templateTypeToFieldName(templateType)

  return {
    name,
    templateType,
    deadline,
    ...(fieldName && { [fieldName]: rest }),
  }
}

export function flowToWorkflow(origin: NodeExperiment, nodesMap: Record<uuid, NodeExperiment>, edges: Edge[]) {
  const sourceMap = edgesToSourceMap(edges)
  const scannedNodes: uuid[] = []
  const realNexts: uuid[] = []

  function genTemplates(origin: NodeExperiment, level: number): Template[] {
    if (scannedNodes.includes(origin.id)) {
      return []
    }

    scannedNodes.push(origin.id)

    const eds = sourceMap.get(origin.id)
    let nextNodes: NodeExperiment[] = []
    const extraNodes: Template[] = []

    eds?.forEach((edge) => {
      if (edge.target) {
        nextNodes.push(nodesMap[edge.target])
      }
    })

    // This indicates that the next node is parallel.
    if (nextNodes.length > 1) {
      extraNodes.push({
        level,
        name: SpecialTemplateType.Parallel + '-' + uuidv4(),
        templateType: SpecialTemplateType.Parallel,
        children: nextNodes.map((n) => n.name),
      })

      let realNext: uuid = ''
      const uniqNexts = _.uniqWith(
        nextNodes.map((n) => {
          const nds = findNextNodeArray(n.id, [], sourceMap)

          return { ...n, next: nds }
        }),
        (a, b) => {
          const intersection = _.intersection<uuid>(a.next, b.next)

          if (intersection.length > 0) {
            realNext = intersection[0]
          }

          return a.next[0] === b.next[0]
        }
      )
      // If all next nodes have the same next node, then jump to the next node.
      const sameNext = uniqNexts.length === 1 && uniqNexts[0] && nodesMap[realNext]

      if (sameNext) {
        nextNodes.forEach((n) => {
          extraNodes.push({ level: level + 1, ...nodeExperimentToTemplate(n) })
        })

        nextNodes = [sameNext]
      }

      // This indicates that all next nodes have non-direct next node.
      if (realNext && !sameNext) {
        realNexts.push(realNext)
      }
    }

    return [
      { level, ...nodeExperimentToTemplate(origin) },
      ...extraNodes,
      ...nextNodes.flatMap((node) =>
        genTemplates(
          node,
          nextNodes.length > 1
            ? level + 1
            : nextNodes.length === 1 && realNexts.includes(nextNodes[0].id)
            ? level - 1
            : level
        )
      ),
    ]
  }

  function findPotentialSerials(nodeName: string, siblings: string[], templates: Template[]) {
    const node = templates.find((t) => t.name === nodeName)!
    let matchedIndex = -1
    const children = []

    for (let i = 0; i < templates.length; i++) {
      const name = templates[i].name

      if (name === nodeName) {
        matchedIndex = i
      }

      if (realNexts.includes(templates[i].id!) || siblings.includes(name) || i === templates.length - 1) {
        return children.length > 1
          ? {
              level: node.level,
              name: SpecialTemplateType.Serial + '-' + uuidv4(),
              templateType: SpecialTemplateType.Serial,
              children,
            }
          : null
      }

      if (matchedIndex > 0 && templates[i].level === node.level && !siblings.includes(name)) {
        children.push(name)
      }
    }
  }

  function genPotentialSerials(templates: Template[]) {
    return templates
      .map((template) => {
        const serials: Template[] = []

        if (template.templateType === SpecialTemplateType.Parallel) {
          template.children = template.children?.map((child, i) => {
            const serial = findPotentialSerials(
              child,
              template.children!.slice(i).filter((name) => name !== child),
              templates
            )

            if (serial) {
              serials.push(serial)

              return serial.name
            }

            return child
          })
        }

        return [template, ...serials]
      })
      .flat()
  }

  let templates = genPotentialSerials(genTemplates(origin, 0))
  templates = [
    {
      name: 'entry',
      templateType: SpecialTemplateType.Serial,
      children: templates.filter((t) => t.level === 0).map((t) => t.name),
    },
    ...templates.map((t) => _.omit(t, 'level')),
  ]

  return yaml.dump(
    {
      apiVersion: 'chaos-mesh.org/v1alpha1',
      kind: 'Workflow',
      metadata: {},
      spec: {
        entry: 'entry',
        templates,
      },
    },
    {
      replacer: (key, value) => {
        if (isDeepEmpty(value)) {
          return undefined
        }

        // field === 'text-text'/'text-label'
        if (_.has(value, 'key0')) {
          if (_.isString(value['key0'].value)) {
            return _.values(value).reduce((acc, { key, value: val }) => {
              acc[key] = val

              return acc
            }, {})
          } else {
            return _.values(value).map(({ key, value: val }) => _.zip(_.times(val.length, _.constant(key)), val))
          }
        }

        // Parse labels, annotations, labelSelectors, and annotationSelectors to object
        if (['labels', 'annotations', 'labelSelectors', 'annotationSelectors'].includes(key)) {
          return (value as string[]).reduce<Record<string, string>>((acc, val) => {
            const [k, v] = val.replace(/\s/g, '').split(':')
            acc[k] = v

            return acc
          }, {})
        }

        return value
      },
    }
  )
}
