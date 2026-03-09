/*
 * Copyright 2025 Chaos Mesh Authors.
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
import { create } from 'zustand'
import { combine } from 'zustand/middleware'

// TODO: remove below code.
export type TemplateExperiment = any

export interface Branch {
  target: string
  expression: string
}

export interface TemplateCustom {
  container: {
    name: string
    image: string
    command: string[]
  }
  conditionalBranches: Branch[]
}

export enum TemplateType {
  Single = 'single',
  Serial = 'serial',
  Parallel = 'parallel',
  Suspend = 'suspend',
  Custom = 'custom',
}

export interface Template {
  index?: number
  type: TemplateType
  name: string
  deadline?: string
  experiment?: TemplateExperiment
  children?: Template[]
  custom?: TemplateCustom
}
// TODO: remove above code.

export type NodeExperiment = any

export const useWorkflowStore = create(
  combine(
    {
      nodes: {} as Record<uuid, NodeExperiment>,
    },
    (set) => ({
      actions: {
        importNodes: (nodes: Record<uuid, NodeExperiment>) => {
          set({ nodes })
        },
        updateWorkflowNode: (node: NodeExperiment) => {
          set((state) => ({
            nodes: { ...state.nodes, [node.name]: node },
          }))
        },
        removeWorkflowNode: (id: uuid) => {
          set((state) => {
            const newNodes = { ...state.nodes }
            delete newNodes[id]
            return { nodes: newNodes }
          })
        },
        resetWorkflow: () => set({ nodes: {} }),
      },
    }),
  ),
)

export const useWorkflowActions = () => useWorkflowStore((state) => state.actions)
