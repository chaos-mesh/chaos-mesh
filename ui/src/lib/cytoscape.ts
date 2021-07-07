import { Node, WorkflowSingle } from 'api/workflows.type'
import cytoscape, { EdgeDefinition, EventHandler, NodeDefinition, Stylesheet } from 'cytoscape'

import { Theme } from '@material-ui/core'
import _flattenDeep from 'lodash.flattendeep'
import dagre from 'cytoscape-dagre'

cytoscape.use(dagre)

type RecursiveNodeDefinition = NodeDefinition | Array<string | RecursiveNodeDefinition>

function generateWorkflowNodes(detail: WorkflowSingle) {
  const { entry, topology } = detail
  let entryNode: Node
  const nodeMap = new Map(
    topology.nodes.map((n) => {
      if (n.template === entry) {
        entryNode = n
      }

      return [n.name, n]
    })
  )
  function toCytoscapeNode(node: Node): RecursiveNodeDefinition {
    const { name, type, state, template } = node

    if (type === 'SerialNode' && node.serial!.children.length) {
      return [
        type,
        node.serial!.children.filter((d) => d.name).map((d) => toCytoscapeNode(nodeMap.get(d.name)!)),
        node.name,
      ]
    } else if (type === 'ParallelNode' && node.parallel!.children.length) {
      return [
        type,
        node.parallel!.children.filter((d) => d.name).map((d) => toCytoscapeNode(nodeMap.get(d.name)!)),
        node.name,
      ]
    } else if (type === 'TaskNode' && node.conditional_branches!.length) {
      return [
        type,
        node.conditional_branches!.filter((d) => d.name).map((d) => toCytoscapeNode(nodeMap.get(d.name)!)),
        node.name,
      ]
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

  return [toCytoscapeNode(entryNode!)]
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
  nodes: RecursiveNodeDefinition[]
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

          // eslint-disable-next-line no-loop-func
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
  onNodeClick: EventHandler
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
  }

  updateElements(detail)

  return { updateElements }
}
