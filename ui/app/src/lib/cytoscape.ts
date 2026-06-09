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
import { Node, WorkflowSingle } from '@/api/workflows.type'
import { Theme } from '@mui/material'
import cytoscape, { EdgeDefinition, EventHandler, NodeDefinition, Stylesheet } from 'cytoscape'
import dagre from 'cytoscape-dagre'
import _ from 'lodash'

cytoscape.use(dagre)

type RecursiveNodeDefinition = NodeDefinition | Array<string | RecursiveNodeDefinition>

function generateWorkflowNodes(detail: WorkflowSingle) {
  const { entry, topology } = detail
  // `nodes` is a nil-able Go slice: a nil slice serializes to `null` (rather than `[]`), so
  // guard the length check with `?.`.
  if (!topology.nodes?.length) {
    return []
  }
  let entryNode: Node | undefined
  const nodeMap = new Map(
    topology.nodes.map((n) => {
      if (n.template === entry) {
        entryNode = n
      }

      return [n.name, n]
    }),
  )
  // Resolve child references to their nodes, dropping any that aren't present (e.g. branches not
  // yet spawned, so their name is still empty, or stale references) so the recursion below never
  // receives an undefined node. Keeping the lookup and the guard together avoids relying on a
  // non-null assertion that a separate filter would have to keep in sync.
  const childNodes = (children: Array<{ name?: string | null }>): RecursiveNodeDefinition[] =>
    children.flatMap((d) => {
      if (!d.name) {
        return []
      }
      const child = nodeMap.get(d.name)
      return child ? [toCytoscapeNode(child)] : []
    })
  function toCytoscapeNode(node: Node): RecursiveNodeDefinition {
    const { name, type, state, template } = node

    // `serial` / `parallel` / `conditional_branches` are nil-able Go slices and arrive as `null`
    // for the node types that don't use them, so normalize to an array before using them.
    const serial = node.serial ?? []
    const parallel = node.parallel ?? []
    const conditionalBranches = node.conditional_branches ?? []

    if (type === 'SerialNode' && serial.length) {
      return [type, childNodes(serial), node.name]
    } else if (type === 'ParallelNode' && parallel.length) {
      return [type, childNodes(parallel), node.name]
    } else if (type === 'TaskNode' && conditionalBranches.length) {
      return [type, childNodes(conditionalBranches), node.name]
    } else {
      return {
        data: {
          id: name,
          type,
          state,
          template,
        },
        classes: state,
      }
    }
  }

  if (!entryNode) {
    return []
  }

  return [toCytoscapeNode(entryNode)]
}

function mergeStates(nodes: NodeDefinition[]) {
  if (nodes.every((n) => n.data.state === 'Succeed')) {
    return 'Succeed'
  }

  return undefined
}

// function connectSerial(edges: EdgeDefinition[], id: string, serial: RecursiveNodeDefinition[]) {
//   const length = serial.length
//   const first = serial[0]
//   const last = serial[length - 1]

//   if (!Array.isArray(first) && !Array.isArray(last)) {
//     edges.push({
//       data: {
//         id,
//         source: first.data.id!,
//         target: last.data.id!,
//       },
//       classes: 'bezier',
//     })
//   }
// }

function generateWorkflowEdges(
  result: EdgeDefinition[],
  connections: NodeDefinition[],
  nodes: RecursiveNodeDefinition[],
) {
  let source = nodes[0] as NodeDefinition

  // source != single node
  if (Array.isArray(source)) {
    const type = source[0]

    if (type === 'SerialNode') {
      generateWorkflowEdges(result, connections, [...source[1], ...nodes.slice(1)])

      // connectSerial(result, source[2], source[1])
    } else if (type === 'ParallelNode' || type === 'TaskNode') {
      const c = {
        data: {
          id: `parallel-connection-${source[2]}`,
        },
        classes: 'connection',
      }

      ;(source[1] as NodeDefinition[]).forEach((d) => {
        if (nodes.length >= 1) {
          generateWorkflowEdges(result, connections, [d, c])
        }
      })

      connections.push(c)

      generateWorkflowEdges(result, connections, [c, ...nodes.slice(1)])
    }
  } else {
    // source = single node
    const _nodes = nodes.slice(1)

    for (let i = 0; i < _nodes.length; i++) {
      const sourceID = source.data.id!
      const n = _nodes[i]

      // N (target) != single node
      if (Array.isArray(n)) {
        const target = n
        const type = target[0]

        if (type === 'SerialNode') {
          generateWorkflowEdges(result, connections, [
            source,
            ...(target[1] as RecursiveNodeDefinition[]),
            ...nodes.slice(i + 2),
          ])

          // connectSerial(result, target[2] as string, target[1] as NodeDefinition[])

          break
        } else if (type === 'ParallelNode') {
          const c1 = {
            data: {
              id: `parallel-connection-${target[2]}`,
            },
            classes: 'connection',
          }

          ;(target[1] as NodeDefinition[]).forEach((d) => {
            generateWorkflowEdges(result, connections, [source, d])
            generateWorkflowEdges(result, connections, [d, c1])
          })

          generateWorkflowEdges(result, connections, [c1, ...nodes.slice(i + 2)])

          connections.push(c1)

          break
        }
      } else {
        // N (target) = single node
        result.push({
          data: {
            id: `${sourceID}-to-${n.data.id!}`,
            source: sourceID,
            target: n.data.id!,
          },
          classes: mergeStates([source, n]),
        })

        source = n
      }
    }
  }
}

export const constructWorkflowTopology = (
  container: HTMLElement,
  detail: WorkflowSingle,
  theme: Theme,
  onNodeClick: EventHandler,
) => {
  const workflowNodeStyle = {
    width: 24,
    height: 24,
    color: theme.palette.text.primary,
    opacity: 0,
    'background-color': theme.palette.grey[500],
    'text-margin-y': '-12px',
    'text-opacity': 0.87,
    label: 'data(id)',
  }

  const workflowStyle: Stylesheet[] = [
    {
      selector: 'node',
      style: workflowNodeStyle,
    },
    {
      selector: 'node.Succeed',
      style: {
        'background-color': theme.palette.success.main,
      },
    },
    {
      selector: 'node.connection',
      style: {
        content: '',
      },
    },
    {
      selector: 'edge',
      style: {
        width: 3,
        opacity: 0,
        'line-color': theme.palette.grey[500],
        'line-opacity': 0.38,
        'curve-style': 'taxi',
        'taxi-turn': '50%',
      } as any,
    },
    {
      selector: 'edge.Succeed',
      style: {
        'line-color': theme.palette.success.main,
      },
    },
    {
      selector: 'edge.bezier',
      style: {
        'curve-style': 'bezier',
        'control-point-step-size': 40,
      },
    },
  ]

  function generateElements(detail: WorkflowSingle) {
    let nodes = generateWorkflowNodes(detail)!
    const edges = [] as EdgeDefinition[]
    const connections = [] as NodeDefinition[]
    generateWorkflowEdges(edges, connections, nodes)
    nodes = nodes.concat(connections)

    return {
      nodes: _.flattenDeep(nodes).filter((d) => typeof d !== 'string') as NodeDefinition[],
      edges: edges,
    }
  }

  const layout = {
    name: 'dagre',
    fit: false,
    rankDir: 'LR',
    nodeSep: 250,
    minLen: 9,
  } as any

  const animateOptions = (style: any) => ({
    style,
    easing: 'ease-in-out' as const,
  })

  const cy = cytoscape({
    container,
    style: workflowStyle,
    minZoom: 0.5,
    maxZoom: 1.5,
  })
    .pan({ x: 250, y: 250 })
    .zoom(0.75)
    .on('click', 'node', function (e) {
      if (e.target.hasClass('connection')) {
        return
      }

      onNodeClick(e)
    })

  let flashRunning: number
  function updateElements(detail: WorkflowSingle) {
    clearInterval(flashRunning)
    flashRunning = window.setInterval(() => {
      const nodes = cy.$('node.Running')

      if (nodes.length) {
        nodes
          .animate(animateOptions({ 'background-opacity': 0.12 }), { duration: 750 })
          .animate(animateOptions({ 'background-opacity': 1 }), { duration: 750 })
      } else {
        clearInterval(flashRunning)
      }
    }, 2000)

    const elements = generateElements(detail)
    cy.json({
      elements,
    })
    cy.layout(layout).run()

    cy.elements().animate(animateOptions({ opacity: 1 }), { duration: 500 })
    cy.center()
  }

  updateElements(detail)

  return { updateElements }
}
