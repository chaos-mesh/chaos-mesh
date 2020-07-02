import { createAsyncThunk, createSlice } from '@reduxjs/toolkit'

import { ExperimentsAction } from './experiments.type'
import { StateOfExperiments } from 'api/experiments.type'
import api from 'api'

const defaultExperiments = {
  total: 0,
  running: 0,
  waiting: 0,
  paused: 0,
  failed: 0,
  finished: 0,
}

export const getStateofExperiments = createAsyncThunk(
  'experiments/state',
  async () => (await api.experiments.state()).data
)

export const getNamespaces = createAsyncThunk('common/namespaces', async () => (await api.common.namespaces()).data)
export const getLabels = createAsyncThunk(
  'common/labels',
  async (podNamespaceList: string) => (await api.common.labels(podNamespaceList)).data
)
export const getAnnotations = createAsyncThunk(
  'common/annotations',
  async (podNamespaceList: string) => (await api.common.annotations(podNamespaceList)).data
)

const initialState: {
  namespaces: string[]
  labels: { [key: string]: string[] }
  annotations: { [key: string]: string[] }
  stateOfExperiments: StateOfExperiments
  needToRefreshExperiments: boolean
} = {
  namespaces: [],
  labels: {},
  annotations: {},
  stateOfExperiments: defaultExperiments,
  needToRefreshExperiments: false,
}

const experimentsSlice = createSlice({
  name: 'experiments',
  initialState,
  reducers: {
    setNeedToRefreshExperiments(state, action) {
      state.needToRefreshExperiments = action.payload
    },
  },
  extraReducers: (builder) => {
    builder.addCase(getStateofExperiments.fulfilled, (state, action: ExperimentsAction) => {
      state.stateOfExperiments = action.payload as StateOfExperiments
    })
    builder.addCase(getNamespaces.fulfilled, (state, action: ExperimentsAction) => {
      state.namespaces = action.payload as string[]
    })
    builder.addCase(getLabels.fulfilled, (state, action: ExperimentsAction) => {
      state.labels = action.payload as { [key: string]: string[] }
    })
    builder.addCase(getAnnotations.fulfilled, (state, action: ExperimentsAction) => {
      state.annotations = action.payload as { [key: string]: string[] }
    })
  },
})

export const { setNeedToRefreshExperiments } = experimentsSlice.actions

export default experimentsSlice.reducer
