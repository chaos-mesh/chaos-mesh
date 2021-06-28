import { PayloadAction, createSlice } from '@reduxjs/toolkit'

export type TemplateExperiment = {
  target: any
  basic: any
}
export interface Template {
  index?: number
  type: 'single' | 'serial' | 'parallel' | 'suspend'
  name: string
  deadline?: string
  experiment?: TemplateExperiment
  children?: Template[]
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
