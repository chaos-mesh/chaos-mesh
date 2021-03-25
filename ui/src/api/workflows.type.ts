export interface workflowParams {
  namespace?: string
  status?: 'initializing' | 'Running' | 'Errored' | 'Finished'
}

export interface Workflow {
  name: string
  namespace: string
  entry: string
}

interface MultiNode {
  tasks: string[]
}
type SerialNode = MultiNode
type ParallelNode = MultiNode

export interface Node {
  name: string
  type: string
  state: string
  template: string
  serial?: SerialNode
  parallel?: ParallelNode
}

export interface WorkflowDetail extends Workflow {
  current_nodes: Node[]
  topology: {
    nodes: Node[]
  }
}
