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

export interface Template {
  index?: number
  type: 'single' | 'serial' | 'parallel' | 'suspend' | 'custom'
  name: string
  deadline?: string
  experiment?: TemplateExperiment
  children?: Template[]
  custom?: TemplateCustom
}

const initialState: {
  templates: Template[]
} = {
  templates: [],
}

const workflowsSlice = createSlice({
  name: 'workflows',
  initialState,
  reducers: {
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
    resetWorkflow(state) {
      state.templates = []
    },
  },
})

export const { setTemplate, updateTemplate, deleteTemplate, resetWorkflow } = workflowsSlice.actions

export default workflowsSlice.reducer
