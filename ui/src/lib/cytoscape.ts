import { Node, WorkflowSingle } from 'api/workflows.type'
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
  {
    selector: 'edge.bezier',
    style: {
      'curve-style': 'bezier',
      'control-point-step-size': 40,
    },
  },
]

type RecursiveNodeDefinition = NodeDefinition | Array<string | RecursiveNodeDefinition>

function generateWorkflowNodes(detail: WorkflowSingle) {
  const { entry, topology } = detail
  const nodeMap = new Map(topology.nodes.map((n) => [n.name, n]))
  const entryNode = topology.nodes.find((n) => n.template === entry)
  const mainChildren = entryNode?.serial?.children

  function toCytoscapeNode(node: Node): RecursiveNodeDefinition {
    const { name, type, state, template } = node

    if (type === 'SerialNode' && node.serial!.children.length) {
      return [
        type,
        node.serial!.children.filter((d) => d.name).map((d) => toCytoscapeNode(nodeMap.get(d.name)!)),
        node.name,
      ]
    } else if (type === 'ParallelNode' && node.parallel!.children.length) {
      return [type, node.parallel!.children.map((d) => toCytoscapeNode(nodeMap.get(d.name)!)), node.name]
    } else {
      return {
        data: {
          id: name,
          type,
          state,
          template,
        },
        classes: state,
        grabbable: false,
      }
    }
  }

  return mainChildren!
    .map((d) => nodeMap.get(d.name))
    .filter((d) => d !== undefined)
    .map((d) => toCytoscapeNode(d!))
}

function mergeStates(nodes: NodeDefinition[]) {
  if (nodes.every((n) => n.data.state === 'Succeed')) {
    return 'Succeed'
  }

  return undefined
}

function connectSerial(edges: EdgeDefinition[], id: string, serial: RecursiveNodeDefinition[]) {
  const length = serial.length
  const first = serial[0]
  const last = serial[length - 1]

  if (!Array.isArray(first) && !Array.isArray(last)) {
    edges.push({
      data: {
        id,
        source: first.data.id!,
        target: last.data.id!,
      },
      classes: 'bezier',
    })
  }
}

function generateWorkflowEdges(result: EdgeDefinition[], nodes: RecursiveNodeDefinition[]) {
  let source = nodes[0] as NodeDefinition

  // source != single node
  if (Array.isArray(source)) {
    const type = source[0]

    if (type === 'SerialNode') {
      generateWorkflowEdges(result, [...source[1], ...nodes.slice(1)])

      // connectSerial(result, source[2], source[1])
    } else if (type === 'ParallelNode') {
      ;(source[1] as NodeDefinition[]).forEach((d) => {
        if (nodes.length > 1) {
          generateWorkflowEdges(result, [d, nodes[1]])
        }
      })

      generateWorkflowEdges(result, nodes.slice(1))
    }
  } else {
    // source = single node
    const nodes_ = nodes.slice(1)

    for (let i = 0; i < nodes_.length; i++) {
      const sourceID = source.data.id!
      const n = nodes_[i]

      // N (target) != single node
      if (Array.isArray(n)) {
        const target = n
        const type = target[0]

        if (type === 'SerialNode') {
          generateWorkflowEdges(result, [source, ...(target[1] as RecursiveNodeDefinition[]), ...nodes.slice(i + 2)])

          // connectSerial(result, target[2] as string, target[1] as NodeDefinition[])

          break
        } else if (type === 'ParallelNode') {
          const connection = {
            data: {
              id: `parallel-connection-${i}`,
            },
            style: {
              label: '',
            },
            grabbable: false,
          }

          // eslint-disable-next-line no-loop-func
          ;(target[1] as NodeDefinition[]).forEach((d) => {
            generateWorkflowEdges(result, [source, d])
            generateWorkflowEdges(result, [d, connection])
          })

          generateWorkflowEdges(result, [connection, ...nodes.slice(i + 2)])

          nodes.push(connection)

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
  onNodeClick: EventHandler
) => {
  function generateElements(detail: WorkflowSingle) {
    const nodes = generateWorkflowNodes(detail)!
    const edges = [] as EdgeDefinition[]
    generateWorkflowEdges(edges, nodes)

    return {
      nodes: _flattenDeep(nodes).filter((d) => typeof d !== 'string') as NodeDefinition[],
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
  }

  updateElements(detail)

  return { updateElements }
}
