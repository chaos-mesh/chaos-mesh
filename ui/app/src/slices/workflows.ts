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
import { PayloadAction, createSlice } from '@reduxjs/toolkit'

import LS from 'lib/localStorage'

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

// TODO: remove above code
export type NodeExperiment = any

export interface WorkflowNode {
  id: uuid
  experiment?: NodeExperiment
}

export interface RecentUse {
  kind: string
  act?: string
}

const initialState: {
  nodes: Record<uuid, NodeExperiment>
  recentUse: RecentUse[]
  templates: Template[]
} = {
  nodes: {},
  recentUse: [],
  templates: [],
}

const workflowSlice = createSlice({
  name: 'workflows',
  initialState,
  reducers: {
    importNodes(state, action: PayloadAction<Record<uuid, NodeExperiment>>) {
      state.nodes = action.payload
    },
    updateWorkflowNode(state, action) {
      const payload = action.payload

      state.nodes[payload.name] = payload
    },
    removeWorkflowNode(state, action: PayloadAction<uuid>) {
      delete state.nodes[action.payload]
    },
    loadRecentlyUsedExperiments(state) {
      state.recentUse = LS.getObj('new-workflow-recently-used-experiments')
    },
    setRecentlyUsedExperiments(state, action: PayloadAction<RecentUse>) {
      const exp = action.payload

      state.recentUse = [...state.recentUse, exp]

      LS.setObj('new-workflow-recently-used-experiments', state.recentUse)
    },
    resetWorkflow(state) {
      state.nodes = {}
      // TODO: remove below code
      state.templates = []
    },
    // TODO: remove below code
    setTemplate(state, action: PayloadAction<Template>) {
      const tpl = action.payload

      state.templates.push(tpl)
    },
    updateTemplate(state, action: PayloadAction<Template>) {
      const { index } = action.payload

      state.templates[index!] = action.payload
    },
    deleteTemplate(state, action: PayloadAction<number>) {
      const index = action.payload
      const templates = state.templates

      state.templates = templates.filter((_, i) => i !== index)
    },
  },
})

export const {
  importNodes,
  updateWorkflowNode,
  removeWorkflowNode,
  loadRecentlyUsedExperiments,
  setRecentlyUsedExperiments,
  resetWorkflow,
  setTemplate,
  updateTemplate,
  deleteTemplate,
} = workflowSlice.actions

export default workflowSlice.reducer
