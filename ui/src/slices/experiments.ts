import { createAsyncThunk, createSlice } from '@reduxjs/toolkit'

import { ExperimentScope } from 'components/NewExperiment/types'
import { ExperimentsAction } from './experiments.type'
import { StateOfExperiments } from 'api/experiments.type'
import api from 'api'

const defaultExperiments = {
  Total: 0,
  Running: 0,
  Waiting: 0,
  Paused: 0,
  Failed: 0,
  Finished: 0,
}

export const getStateofExperiments = createAsyncThunk(
  'experiments/state',
  async () => (await api.experiments.state()).data
)

export const getNamespaces = createAsyncThunk('common/chaos-available-namespaces', async () => (await api.common.chaosAvailableNamespaces()).data)
export const getLabels = createAsyncThunk(
  'common/labels',
  async (podNamespaceList: string[]) => (await api.common.labels(podNamespaceList)).data
)
export const getAnnotations = createAsyncThunk(
  'common/annotations',
  async (podNamespaceList: string[]) => (await api.common.annotations(podNamespaceList)).data
)
export const getPodsByNamespaces = createAsyncThunk(
  'common/pods',
  async (data: Partial<ExperimentScope>) => (await api.common.pods(data)).data
)

const initialState: {
  namespaces: string[]
  labels: Record<string, string[]>
  annotations: Record<string, string[]>
  pods: any[]
  stateOfExperiments: StateOfExperiments
  needToRefreshExperiments: boolean
} = {
  namespaces: [],
  labels: {},
  annotations: {},
  pods: [],
  stateOfExperiments: defaultExperiments,
  needToRefreshExperiments: false,
}

const namespaceFilters = ['kube-system', 'chaos-testing']

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
      state.namespaces = (action.payload as string[]).filter((d) => !namespaceFilters.includes(d))
    })
    builder.addCase(getLabels.fulfilled, (state, action: ExperimentsAction) => {
      state.labels = action.payload as Record<string, string[]>
    })
    builder.addCase(getAnnotations.fulfilled, (state, action: ExperimentsAction) => {
      state.annotations = action.payload as Record<string, string[]>
    })
    builder.addCase(getPodsByNamespaces.fulfilled, (state, action) => {
      state.pods = action.payload as any[]
    })
  },
})

export const { setNeedToRefreshExperiments } = experimentsSlice.actions

export default experimentsSlice.reducer
