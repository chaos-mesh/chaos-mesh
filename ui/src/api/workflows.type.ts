export interface WorkflowParams {
  namespace?: string
}

export interface Workflow {
  uid: uuid
  namespace: string
  name: string
  entry: string
  created_at: string
  end_time: string
  status: 'Succeed'
}

interface MultiNode {
  tasks: { name: string; template: string }[]
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

export interface WorkflowSingle extends Workflow {
  topology: {
    nodes: Node[]
  }
  kube_object: any
}
