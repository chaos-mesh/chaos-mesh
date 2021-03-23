import cytoscape, { CytoscapeOptions, Stylesheet } from 'cytoscape'

import dagre from 'cytoscape-dagre'

cytoscape.use(dagre)

export const workflowStyle: Stylesheet[] = [
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
    } as any,
  },
]

export default function constructCytoscape(container: HTMLElement, options: CytoscapeOptions) {
  return cytoscape({
    container,
    ...options,
  })
}
