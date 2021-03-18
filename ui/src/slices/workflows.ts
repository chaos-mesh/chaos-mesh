import { PayloadAction, createSlice } from '@reduxjs/toolkit'

import { Experiment } from 'components/NewExperiment/types'

interface Template {
  type: 'single' | 'serial' | 'parallel'
  experiment: Experiment
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
      state.templates = [...state.templates, action.payload]
    },
  },
})

export const { setTemplate } = workflowsSlice.actions

export default workflowsSlice.reducer
