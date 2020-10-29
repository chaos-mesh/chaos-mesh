import { PayloadAction, createAsyncThunk, createSlice } from '@reduxjs/toolkit'

import { ExperimentScope } from 'components/NewExperiment/types'
import { Kind } from 'components/NewExperimentNext/data/target'
import { StateOfExperiments } from 'api/experiments.type'
import api from 'api'

const defaultStateOfExperiments = {
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

export const getNamespaces = createAsyncThunk('common/namespaces', async () => (await api.common.namespaces()).data)
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
  step1: boolean
  step2: boolean
  kind: Kind | ''
  action: string
  target: any
  basic: any
} = {
  namespaces: [],
  labels: {},
  annotations: {},
  pods: [],
  stateOfExperiments: defaultStateOfExperiments,
  needToRefreshExperiments: false,
  // New Experiment needed
  step1: false,
  step2: false,
  kind: '',
  action: '',
  target: {},
  basic: {},
}

const namespaceFilters = ['kube-system']

const experimentsSlice = createSlice({
  name: 'experiments',
  initialState,
  reducers: {
    setNeedToRefreshExperiments(state, action: PayloadAction<boolean>) {
      state.needToRefreshExperiments = action.payload
    },
    setStep1(state, action: PayloadAction<boolean>) {
      state.step1 = action.payload
    },
    setStep2(state, action: PayloadAction<boolean>) {
      state.step2 = action.payload
    },
    setKind(state, action) {
      state.kind = action.payload
    },
    setAction(state, action) {
      state.action = action.payload
    },
    setTarget(state, action) {
      state.target = action.payload
    },
    setBasic(state, action) {
      state.basic = action.payload
    },
  },
  extraReducers: (builder) => {
    builder.addCase(getStateofExperiments.fulfilled, (state, action) => {
      state.stateOfExperiments = action.payload
    })
    builder.addCase(getNamespaces.fulfilled, (state, action) => {
      state.namespaces = action.payload.filter((d) => !namespaceFilters.includes(d))
    })
    builder.addCase(getLabels.fulfilled, (state, action) => {
      state.labels = action.payload
    })
    builder.addCase(getAnnotations.fulfilled, (state, action) => {
      state.annotations = action.payload
    })
    builder.addCase(getPodsByNamespaces.fulfilled, (state, action) => {
      state.pods = action.payload as any[]
    })
  },
})

export const {
  setNeedToRefreshExperiments,
  setStep1,
  setStep2,
  setKind,
  setAction,
  setTarget,
  setBasic,
} = experimentsSlice.actions

export default experimentsSlice.reducer
