import { PayloadAction, createSlice } from '@reduxjs/toolkit'

export interface TemplateExperiment {
  target: any
  basic: any
}

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
