import { Node, WorkflowDetail } from 'api/workflows.type'
import cytoscape, { EdgeDefinition, NodeDefinition, Stylesheet } from 'cytoscape'

import _flattenDeep from 'lodash.flattendeep'
import dagre from 'cytoscape-dagre'

cytoscape.use(dagre)

const workflowStyle: Stylesheet[] = [
  {
    selector: 'node',
    style: {
      width: 16,
      height: 16,
      color: 'rgb(0, 0, 0)',
      'background-color': 'grey',
      'text-margin-y': '-3px',
      'text-opacity': 0.87,
      label: 'data(id)',
    },
  },
  {
    selector: 'edge',
    style: {
      width: 3,
      'line-color': 'rgb(0, 0, 0)',
      'line-opacity': 0.12,
      'curve-style': 'taxi',
      'taxi-direction': 'horizontal',
      'taxi-turn': 100,
    } as any,
  },
]

type RecursiveNodeDefinition = NodeDefinition | Array<string | RecursiveNodeDefinition>

function generateWorkflowNodes(detail: WorkflowDetail) {
  const { entry, topology } = detail
  const nodeMap = new Map(topology.nodes.map((n) => [n.template, n]))
  const entryNode = nodeMap.get(entry)
  const order = entryNode?.serial?.tasks

  function toCytoscapeNode(node: Node): RecursiveNodeDefinition {
    const { template, type } = node

    if (type === 'SerialNode') {
      return [type, node.serial!.tasks.map((d) => toCytoscapeNode(nodeMap.get(d)!))]
    } else if (type === 'ParallelNode') {
      return [type, node.parallel!.tasks.map((d) => toCytoscapeNode(nodeMap.get(d)!))]
    } else {
      return {
        data: {
          id: template,
          type,
        },
      }
    }
  }

  if (order) {
    return order.map((d) => nodeMap.get(d)).map((d) => toCytoscapeNode(d!))
  }
}

type RecursiveEdgeDefinition = EdgeDefinition | Array<string | RecursiveEdgeDefinition>

function generateWorkflowEdges(nodes: RecursiveNodeDefinition[]) {
  const result: RecursiveEdgeDefinition[] = []

  nodes.forEach((n, i, arr) => {
    if (i === nodes.length - 1) {
      return
    }

    // N (source) is not a single node
    if (Array.isArray(n)) {
      return generateWorkflowEdges(n[1] as RecursiveNodeDefinition[])
    } else {
      // N (source) is a single node
      const source = n.data.id!
      const target = arr[i + 1]

      // The target is not a single node
      if (Array.isArray(target)) {
        const type = target[0]

        if (type === 'SerialNode') {
          const firstNode = (target[1] as NodeDefinition[])[0]
          const targetID = firstNode.data.id!

          result.push({
            data: {
              id: `${source}-to-${targetID}`,
              source,
              target: targetID,
            },
          })
        } else if (type === 'ParallelNode') {
          ;(target[1] as NodeDefinition[]).map((d) => ({
            data: {
              id: `${source}-to-${d.data.id!}`,
              source,
              target: d.data.id!,
            },
          }))
        }
      } else {
        // The target is a single node
        result.push({
          data: {
            id: `${source}-to-${target.data.id!}`,
            source,
            target: target.data.id!,
          },
        })
      }
    }
  })

  return result
}

export const constructWorkflowTopology = (container: HTMLElement, detail: WorkflowDetail) => {
  const nodes = generateWorkflowNodes(detail)!
  const edges = generateWorkflowEdges(nodes)

  cytoscape({
    container,
    elements: {
      nodes: _flattenDeep(nodes).filter((d) => typeof d !== 'string') as NodeDefinition[],
      edges: edges as EdgeDefinition[],
    },
    style: workflowStyle,
    layout: {
      name: 'dagre',
      rankDir: 'LR',
      nodeSep: 250,
      minLen: 9,
    } as any,
  })
}
