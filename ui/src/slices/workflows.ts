import { PayloadAction, createSlice } from '@reduxjs/toolkit'

export type TemplateExperiment = {
  target: any
  basic: any
}
export interface Template {
  type: 'single' | 'serial' | 'parallel' | 'suspend'
  index?: number
  name: string
  experiments: TemplateExperiment[]
  suspend?: { duration: string }
}

let index = 0

const initialState: {
  templates: Record<string, Template>
} = {
  templates: {},
}

const workflowsSlice = createSlice({
  name: 'workflows',
  initialState,
  reducers: {
    setTemplate(state, action: PayloadAction<Template>) {
      const tpl = action.payload
      const { name } = tpl

      tpl.index = index++
      state.templates[name!] = tpl
    },
    updateTemplate(state, action: PayloadAction<Template>) {
      const { name } = action.payload

      state.templates[name!] = action.payload
    },
    deleteTemplate(state, action: PayloadAction<string>) {
      const name = action.payload
      const templates = state.templates

      delete templates[name]

      state.templates = templates
      index--
    },
  },
})

export const { setTemplate, updateTemplate, deleteTemplate } = workflowsSlice.actions

export default workflowsSlice.reducer
