/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
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

export interface RequestForm {
  name: string
  url: string
  method: string
  body: string
  followLocation: boolean
  jsonContent: boolean
}
