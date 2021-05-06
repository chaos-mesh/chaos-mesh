export interface workflowParams {
  namespace?: string
}

export interface Workflow {
  name: string
  namespace: string
  entry: string
  created: string
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
  topology: {
    nodes: Node[]
  }
  kube_object: any
}
