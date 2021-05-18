import { Node, WorkflowDetail } from 'api/workflows.type'
import cytoscape, { EdgeDefinition, EventHandler, NodeDefinition, Stylesheet } from 'cytoscape'

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
  const nodeMap = new Map(topology.nodes.map((n) => [n.name, n]))
  const entryNode = topology.nodes.find((n) => n.template === entry)
  const mainTasks = entryNode?.serial?.tasks

  function toCytoscapeNode(node: Node): RecursiveNodeDefinition {
    const { name, type, state } = node

    if (type === 'SerialNode' && node.serial!.tasks.length) {
      return [type, node.serial!.tasks.map((d) => toCytoscapeNode(nodeMap.get(d.name)!))]
    } else if (type === 'ParallelNode' && node.parallel!.tasks.length) {
      return [type, node.parallel!.tasks.map((d) => toCytoscapeNode(nodeMap.get(d.name)!))]
    } else {
      return {
        data: {
          id: name,
          type,
          state,
        },
        classes: state,
        grabbable: false,
      }
    }
  }

  return mainTasks!
    .map((d) => nodeMap.get(d.name))
    .filter((d) => d !== undefined)
    .map((d) => toCytoscapeNode(d!))
}

function mergeStates(source: Node['state'], target: Node['state']) {
  if (source === 'Succeed' && target === 'Succeed') {
    return 'Succeed'
  }

  return undefined
}

function generateWorkflowEdges(result: EdgeDefinition[], nodes: RecursiveNodeDefinition[]) {
  let first = nodes[0]

  // source != single node
  if (Array.isArray(first)) {
    generateWorkflowEdges(result, first[1] as RecursiveNodeDefinition[])
  } else {
    // source = single node
    let source = first

    nodes.slice(1).forEach((n) => {
      const sourceID = source.data.id!
      const sourceState = source.data.state

      // N (target) = single node
      if (Array.isArray(n)) {
        const target = n
        const type = target[0]

        if (type === 'SerialNode') {
          const length = (target[1] as NodeDefinition[]).length
          const firstNode = (target[1] as NodeDefinition[])[0]
          const targetID = firstNode.data.id!
          const state = mergeStates(sourceState, firstNode.data.state)

          result.push({
            data: {
              id: `${sourceID}-to-${targetID}`,
              source: sourceID,
              target: targetID,
            },
            classes: state,
          })

          source = (target[1] as NodeDefinition[])[length - 1]

          generateWorkflowEdges(result, target[1] as RecursiveNodeDefinition[])
        } else if (type === 'ParallelNode') {
          ;(target[1] as NodeDefinition[]).forEach((d) => {
            const state = mergeStates(sourceState, d.data.state)

            const edge = {
              data: {
                id: `${sourceID}-to-${d.data.id!}`,
                source: sourceID,
                target: d.data.id!,
              },
              classes: state,
            }

            result.push(edge)
          })
        }
      } else {
        // N (target) != single node
        result.push({
          data: {
            id: `${sourceID}-to-${n.data.id!}`,
            source: sourceID,
            target: n.data.id!,
          },
          classes: mergeStates(sourceState, n.data.state),
        })

        source = n
      }
    })
  }

  return result
}

export const constructWorkflowTopology = (
  container: HTMLElement,
  detail: WorkflowDetail,
  onNodeClick: EventHandler
) => {
  function generateElements(detail: WorkflowDetail) {
    const nodes = generateWorkflowNodes(detail)!
    const edges = generateWorkflowEdges([], nodes)

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
    .on('click', 'node', onNodeClick)

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
      nodes
        .animate(animateOptions({ 'background-opacity': 0.12 }), { duration: 750 })
        .animate(animateOptions({ 'background-opacity': 1 }), { duration: 750 })
    } else {
      clearInterval(flashRunning)
    }
  }, 2000)

  return { updateElements }
}
