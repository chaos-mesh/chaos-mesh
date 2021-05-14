import { Node, WorkflowDetail } from 'api/workflows.type'
import cytoscape, { EdgeDefinition, NodeDefinition, Stylesheet } from 'cytoscape'

import _flattenDeep from 'lodash.flattendeep'
import dagre from 'cytoscape-dagre'
import theme from 'theme'

cytoscape.use(dagre)

const workflowNodeStyle = {
  width: 24,
  height: 24,
  color: 'rgb(0, 0, 0)',
  opacity: 0,
  'background-color': 'grey',
  'text-margin-y': '-3px',
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
    selector: 'edge',
    style: {
      width: 3,
      opacity: 0,
      'line-color': 'rgb(0, 0, 0)',
      'line-opacity': 0.12,
      'curve-style': 'taxi',
      'taxi-direction': 'horizontal',
      'taxi-turn': 100,
    } as any,
  },
  {
    selector: 'edge.Succeed',
    style: {
      'line-color': theme.palette.success.main,
    },
  },
]

type RecursiveNodeDefinition = NodeDefinition | Array<string | RecursiveNodeDefinition>

function generateWorkflowNodes(detail: WorkflowDetail) {
  const { entry, topology } = detail
  const nodeMap = new Map(topology.nodes.map((n) => [n.template, n]))
  const entryNode = nodeMap.get(entry)
  const order = entryNode?.serial?.tasks

  function toCytoscapeNode(node: Node): RecursiveNodeDefinition {
    const { template, type, state } = node

    if (type === 'SerialNode' && node.serial!.tasks.length) {
      return [type, node.serial!.tasks.map((d) => toCytoscapeNode(nodeMap.get(d)!))]
    } else if (type === 'ParallelNode' && node.parallel!.tasks.length) {
      return [type, node.parallel!.tasks.map((d) => toCytoscapeNode(nodeMap.get(d)!))]
    } else {
      return {
        data: {
          id: template,
          type,
          state,
        },
        classes: state,
        grabbable: false,
      }
    }
  }

  if (order) {
    return order
      .map((d) => nodeMap.get(d))
      .filter((d) => d !== undefined)
      .map((d) => toCytoscapeNode(d!))
  }
}

function mergeStates(source: Node['state'], target: Node['state']) {
  if (source === 'Succeed' && target === 'Succeed') {
    return 'Succeed'
  }

  return undefined
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
      const sourceState = n.data.state
      const target = arr[i + 1]

      // The target is not a single node
      if (Array.isArray(target)) {
        const type = target[0]

        if (type === 'SerialNode') {
          const firstNode = (target[1] as NodeDefinition[])[0]
          const targetID = firstNode.data.id!
          const state = mergeStates(sourceState, firstNode.data.state)

          result.push({
            data: {
              id: `${source}-to-${targetID}`,
              source,
              target: targetID,
            },
            classes: state,
          })
        } else if (type === 'ParallelNode') {
          ;(target[1] as NodeDefinition[]).map((d) => {
            const state = mergeStates(sourceState, d.data.state)

            return {
              data: {
                id: `${source}-to-${d.data.id!}`,
                source,
                target: d.data.id!,
              },
              classes: state,
            }
          })
        }
      } else {
        // The target is a single node
        result.push({
          data: {
            id: `${source}-to-${target.data.id!}`,
            source,
            target: target.data.id!,
          },
          classes: mergeStates(sourceState, target.data.state),
        })
      }
    }
  })

  return result
}

export const constructWorkflowTopology = (container: HTMLElement, detail: WorkflowDetail) => {
  function generateElements(detail: WorkflowDetail) {
    const nodes = generateWorkflowNodes(detail)!
    const edges = generateWorkflowEdges(nodes)

    return {
      nodes: _flattenDeep(nodes).filter((d) => typeof d !== 'string') as NodeDefinition[],
      edges: edges as EdgeDefinition[],
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
    easing: 'ease-in-out' as 'ease-in-out',
  })

  const cy = cytoscape({
    container,
    style: workflowStyle,
    minZoom: 0.5,
    maxZoom: 1.5,
  })
    .pan({ x: 150, y: container.offsetHeight / 2 - 50 })
    .zoom(0.75)

  function updateElements(detail: WorkflowDetail) {
    cy.json({
      elements: generateElements(detail),
    })
    cy.layout(layout).run()

    cy.elements().animate(animateOptions({ opacity: 1 }), { duration: 500 })
  }

  updateElements(detail)

  const flashRunning = setInterval(() => {
    const nodes = cy.$('node.Running')

    if (nodes.length) {
      console.log(1)
      nodes
        .animate(animateOptions({ 'background-opacity': 0.12 }), { duration: 750 })
        .animate(animateOptions({ 'background-opacity': 1 }), { duration: 750 })
    } else {
      console.log(2)
      clearInterval(flashRunning)
    }
  }, 2000)

  return { updateElements }
}
