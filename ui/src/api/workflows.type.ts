export interface WorkflowParams {
  namespace?: string
}

export interface Workflow {
  is: 'workflow'
  uid: uuid
  namespace: string
  name: string
  entry: string
  created_at: string
  end_time: string
  status: 'running' | 'finished' | 'failed' | 'unknown'
}

interface MultiNode {
  children: { name: string; template: string }[]
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
