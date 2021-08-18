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

interface NodeNameWithTemplate {
  name: string
  template: string
}
type SerialNode = NodeNameWithTemplate[]
type ParallelNode = NodeNameWithTemplate[]

interface ConditionalBranch {
  name: string
  template: string
  Expression: string
}

export interface Node {
  name: string
  type: 'ChaosNode' | 'SerialNode' | 'ParallelNode' | 'SuspendNode' | 'TaskNode'
  state: string
  template: string
  serial?: SerialNode
  parallel?: ParallelNode
  conditional_branches?: Array<ConditionalBranch>
}

export interface WorkflowSingle extends Workflow {
  topology: {
    nodes: Node[]
  }
  kube_object: any
}
