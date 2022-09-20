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
import { Edge, Node, XYPosition, getIncomers } from 'react-flow-renderer'
import { v4 as uuidv4 } from 'uuid'

import type { NodeExperiment } from 'slices/workflows'

import { Schedule, scheduleInitialValues } from 'components/AutoForm/data'

import { arrToObjBySep, isDeepEmpty, objToArrBySep } from 'lib/utils'

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

export function templateTypeToFieldName(templateType: ExperimentKind) {
  return mapping.get(templateType)
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
  schedule?: { type: string; [key: string]: any } & Schedule
  children?: string[]
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
        [templateTypeToFieldName(templateType)!]: restrest,
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

export function flowToWorkflow(nodes: Node[], edges: Edge[], storeTemplates: Record<string, NodeExperiment>) {
  const origin = nodes
    .filter((n) => !n.parentNode)
    .map((n) => ({ ...n, incomers: getIncomers(n, nodes, edges) }))
    .find((n) => n.incomers.length === 0)!
  const nodeMap = _.keyBy(nodes, 'id')
  const sourceMap = _.keyBy(edges, 'source')

  function genTemplates(origin: Node, level: number): Template[] {
    const originalTemplate = storeTemplates[origin.data.name]
    let currentTemplate: Template
    let restTemplates: Template[] = []

    if (
      originalTemplate.templateType === SpecialTemplateType.Serial ||
      originalTemplate.templateType === SpecialTemplateType.Parallel
    ) {
      const children = nodes
        .filter((n) => n.parentNode === origin.id)
        .map((n) => ({ id: n.id, ...storeTemplates[n.data.name] }))

      currentTemplate = {
        level,
        name: originalTemplate.name,
        templateType: originalTemplate.templateType,
        children: children.map((n) => n.name),
      }

      restTemplates = children.flatMap((n) => genTemplates(nodeMap[n.id], level + 1))
    } else {
      currentTemplate = { level, ...nodeExperimentToTemplate(originalTemplate) }
    }

    const edge = sourceMap[origin.id]
    let nextNode
    if (edge) {
      nextNode = nodeMap[edge.target]
    }

    return [
      currentTemplate,
      ...restTemplates,
      ...(nextNode && !nextNode.parentNode ? genTemplates(nextNode, level) : []),
    ]
  }

  let templates = _.uniqBy(genTemplates(origin, 0), 'name')
  const templatesWithLevel0 = templates.filter((t) => t.level === 0)
  const hasEntry =
    templatesWithLevel0.length === 1 &&
    (templatesWithLevel0[0].templateType === SpecialTemplateType.Serial ||
      templatesWithLevel0[0].templateType === SpecialTemplateType.Parallel)
  templates = [
    ...(!hasEntry
      ? [
          {
            name: 'entry',
            templateType: SpecialTemplateType.Serial,
            children: templatesWithLevel0.map((t) => t.name),
          },
        ]
      : []),
    ...templates.map((t) => _.omit(t, 'level')),
  ]

  return yaml.dump(
    {
      apiVersion: 'chaos-mesh.org/v1alpha1',
      kind: 'Workflow',
      metadata: {},
      spec: {
        entry: hasEntry ? templatesWithLevel0[0].name : 'entry',
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

        // Parse labels, annotations, labelSelectors, and annotationSelectors to object.
        if (['labels', 'annotations', 'labelSelectors', 'annotationSelectors'].includes(key)) {
          return arrToObjBySep(value, ': ')
        }

        return value
      },
    }
  )
}

export function templateToNodeExperiment(t: Template, scheduled?: boolean): NodeExperiment {
  let result

  if (scheduled) {
    const { type, schedule, historyLimit, concurrencyPolicy, startingDeadlineSeconds, ...rest } = t['schedule']!
    const fieldName = templateTypeToFieldName(type as ExperimentKind)!
    const chaos = rest[fieldName]

    result = {
      kind: type,
      name: t.name,
      templateType: type,
      deadline: t.deadline,
      scheduled: true,
      ..._.defaults(
        {
          schedule,
          historyLimit,
          concurrencyPolicy,
          startingDeadlineSeconds,
        },
        scheduleInitialValues
      ),
      ...chaos,
    }
  } else {
    const fieldName = templateTypeToFieldName(t.templateType as ExperimentKind)

    result = {
      kind: t.templateType,
      name: t.name,
      templateType: t.templateType,
      deadline: t.deadline,
      ...(fieldName && (t as any)[fieldName]),
    }
  }

  // Parse labelSelectors, and annotationSelectors to array.
  if (_.has(result, 'selector.labelSelectors')) {
    _.update(result, 'selector.labelSelectors', (obj) => objToArrBySep(obj, ': '))
  }
  if (_.has(result, 'selector.annotationSelectors')) {
    _.update(result, 'selector.annotationSelectors', (obj) => objToArrBySep(obj, ': '))
  }

  return result
}

export function connectNodes(nodes: Node[]) {
  const edges: Edge[] = []

  for (let i = 1; i < nodes.length; i++) {
    const prev = nodes[i - 1]
    const cur = nodes[i]

    const id = uuidv4()

    edges.push({
      id,
      type: 'adjustableEdge',
      source: prev.id,
      target: cur.id,
      data: {
        id,
      },
    })
  }

  return edges
}

type ParentNode = {
  id: uuid
  type: SpecialTemplateType.Serial | SpecialTemplateType.Parallel
}
export enum View {
  NodeWidth = 200,
  NodeHeight = 30,
  PaddingX = 30,
  PaddingY = 15,
  GroupNodeTypographyHeight = 32,
}

export function workflowToFlow(workflow: string) {
  const { entry, templates }: { entry: string; templates: Template[] } = (yaml.load(workflow) as any).spec
  const templatesMap = _.keyBy(templates, 'name')
  // Convert templates to store.
  //
  // The `name` is used here as the unique id,
  // because the name of a template inside a Workflow is unique.
  const store = _.transform<Template, Record<string, NodeExperiment>>(templatesMap, (acc, t, k) => {
    if (t.templateType === 'Schedule') {
      acc[k] = templateToNodeExperiment(t, true)
    } else {
      acc[k] = templateToNodeExperiment(t)
    }
  })
  const nodes: Record<uuid, Node> = {}
  const edges: Edge[] = []

  function recurInsertNodesAndEdges(
    entry: Template,
    relativePos: XYPosition,
    level: number,
    index: number,
    parentNode?: ParentNode
  ): { id: uuid; width: number; height: number } {
    function addNode(id: uuid, parentNode?: ParentNode): Node {
      return {
        id,
        type: 'flowNode',
        position: {
          x:
            parentNode?.type === SpecialTemplateType.Serial
              ? relativePos.x + View.PaddingX * (index + 1)
              : View.PaddingX,
          y:
            View.GroupNodeTypographyHeight +
            (parentNode?.type === SpecialTemplateType.Parallel
              ? relativePos.y + View.PaddingY * (index + 1)
              : View.PaddingY),
        },
        data: {
          name: entry.name,
          kind: entry.templateType,
          children: _.truncate(entry.name, { length: 20 }),
        },
        ...(parentNode && {
          parentNode: parentNode.id,
          extent: 'parent',
        }),
      }
    }

    const id = uuidv4()
    let width = 0
    let height = 0

    if (entry.templateType === SpecialTemplateType.Serial || entry.templateType === SpecialTemplateType.Parallel) {
      const childrenNum = entry.children!.length

      nodes[id] = {
        id,
        type: 'groupNode',
        position: {
          x:
            parentNode?.type === SpecialTemplateType.Serial
              ? relativePos.x + View.PaddingX * (index + 1)
              : parentNode?.type === SpecialTemplateType.Parallel
              ? View.PaddingX
              : relativePos.x,
          y:
            View.GroupNodeTypographyHeight +
            (parentNode?.type === SpecialTemplateType.Parallel
              ? relativePos.y + View.PaddingY * (index + 1)
              : parentNode?.type === SpecialTemplateType.Serial
              ? View.PaddingY
              : relativePos.y - View.GroupNodeTypographyHeight),
        },
        data: {
          name: entry.name,
          type: entry.templateType,
          childrenNum,
        },
        ...(parentNode && {
          parentNode: parentNode.id,
          extent: 'parent',
          connectable: parentNode.type === SpecialTemplateType.Serial,
        }),
        zIndex: -1, // Make edges visible on the top of the group node.
      }

      const children = entry.children!.map((child) => templatesMap[child])

      let prevWidth = 0
      let prevHeight = 0
      const uuids = children.map((child, i) => {
        const {
          id: uuid,
          width: w,
          height: h,
        } = recurInsertNodesAndEdges(
          child,
          {
            x: prevWidth,
            y: prevHeight,
          },
          level + 1,
          i,
          {
            id,
            type: entry.templateType as any,
          }
        )

        if (entry.templateType === SpecialTemplateType.Serial) {
          width += w
          height = Math.max(height, h)

          prevWidth += w
        }

        if (entry.templateType === SpecialTemplateType.Parallel) {
          width = Math.max(width, w)
          height += h

          prevHeight += h
        }

        if (nodes[uuid].type === 'groupNode') {
          prevHeight += View.GroupNodeTypographyHeight
        }

        return uuid
      })

      // If Serial, connect all child nodes.
      if (entry.templateType === SpecialTemplateType.Serial) {
        edges.push(...connectNodes(uuids.map((uuid) => nodes[uuid])))
      }

      // Calculate the padding of the group node.
      width += entry.templateType === SpecialTemplateType.Serial ? View.PaddingX * (childrenNum + 1) : View.PaddingX * 2
      height +=
        entry.templateType === SpecialTemplateType.Parallel ? View.PaddingY * (childrenNum + 1) : View.PaddingY * 2

      // Calculate the height of all headers.
      const specialUUIDs = _.sumBy(uuids, (uuid) => (nodes[uuid].type === 'groupNode' ? 1 : 0))
      if (specialUUIDs > 0) {
        height +=
          entry.templateType === SpecialTemplateType.Serial
            ? View.GroupNodeTypographyHeight
            : View.GroupNodeTypographyHeight * specialUUIDs
      }

      nodes[id].data.width = width
      nodes[id].data.height = height
    } else {
      nodes[id] = addNode(id, parentNode)

      width = View.NodeWidth
      height = View.NodeHeight
    }

    return { id, width, height }
  }

  recurInsertNodesAndEdges(templatesMap[entry], { x: 100, y: 100 }, 0, 0)

  return { store, nodes: _.values(nodes), edges }
}
