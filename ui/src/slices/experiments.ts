import { PayloadAction, createAsyncThunk, createSlice } from '@reduxjs/toolkit'

import { ExperimentScope } from 'components/NewExperiment/types'
import { Kind } from 'components/NewExperimentNext/data/target'
import api from 'api'

export const getNamespaces = createAsyncThunk(
  'common/chaos-available-namespaces',
  async () => (await api.common.chaosAvailableNamespaces()).data
)
export const getLabels = createAsyncThunk(
  'common/labels',
  async (podNamespaceList: string[]) => (await api.common.labels(podNamespaceList)).data
)
export const getAnnotations = createAsyncThunk(
  'common/annotations',
  async (podNamespaceList: string[]) => (await api.common.annotations(podNamespaceList)).data
)
export const getCommonPodsByNamespaces = createAsyncThunk(
  'common/pods',
  async (data: Partial<ExperimentScope>) => (await api.common.pods(data)).data
)
export const getNetworkTargetPodsByNamespaces = createAsyncThunk(
  'network/target/pods',
  async (data: Partial<ExperimentScope>) => (await api.common.pods(data)).data
)

const initialState: {
  namespaces: string[]
  labels: Record<string, string[]>
  annotations: Record<string, string[]>
  pods: any[]
  networkTargetPods: any[]
  step1: boolean
  step2: boolean
  kindAction: [Kind | '', string]
  target: any
  basic: any
} = {
  namespaces: [],
  labels: {},
  annotations: {},
  pods: [],
  networkTargetPods: [],
  // New Experiment needed
  step1: false,
  step2: false,
  kindAction: ['', ''],
  target: {},
  basic: {},
}

const namespaceFilters = ['kube-system']

const experimentsSlice = createSlice({
  name: 'experiments',
  initialState,
  reducers: {
    clearNetworkTargetPods(state) {
      state.networkTargetPods = []
    },
    setStep1(state, action: PayloadAction<boolean>) {
      state.step1 = action.payload
    },
    setStep2(state, action: PayloadAction<boolean>) {
      state.step2 = action.payload
    },
    setKindAction(state, action) {
      state.kindAction = action.payload
    },
    setTarget(state, action) {
      state.target = action.payload
    },
    setBasic(state, action) {
      state.basic = action.payload
    },
    resetNewExperiment(state) {
      state.step1 = false
      state.step2 = false
      state.kindAction = ['', '']
      state.target = {}
      state.basic = {}
    },
  },
  extraReducers: (builder) => {
    builder.addCase(getNamespaces.fulfilled, (state, action) => {
      state.namespaces = action.payload.filter((d) => !namespaceFilters.includes(d))
    })
    builder.addCase(getLabels.fulfilled, (state, action) => {
      state.labels = action.payload
    })
    builder.addCase(getAnnotations.fulfilled, (state, action) => {
      state.annotations = action.payload
    })
    builder.addCase(getCommonPodsByNamespaces.fulfilled, (state, action) => {
      state.pods = action.payload as any[]
    })
    builder.addCase(getNetworkTargetPodsByNamespaces.fulfilled, (state, action) => {
      state.networkTargetPods = action.payload as any[]
    })
  },
})

export const {
  clearNetworkTargetPods,
  setStep1,
  setStep2,
  setKindAction,
  setTarget,
  setBasic,
  resetNewExperiment,
} = experimentsSlice.actions

export default experimentsSlice.reducer
