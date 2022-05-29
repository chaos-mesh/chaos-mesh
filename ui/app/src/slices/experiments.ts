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
import { PayloadAction, createAsyncThunk, createSlice } from '@reduxjs/toolkit'
import api from 'api'

import { Kind } from 'components/NewExperimentNext/data/types'

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
export const getCommonPods = createAsyncThunk(
  'common/pods',
  async (data: Record<string, any>) => (await api.common.pods(data)).data
)
export const getNetworkTargetPods = createAsyncThunk(
  'network/target/pods',
  async (data: Record<string, any>) => (await api.common.pods(data)).data
)

export type Env = 'k8s' | 'physic'

const initialState: {
  namespaces: string[]
  labels: Record<string, string[]>
  annotations: Record<string, string[]>
  pods: any[]
  networkTargetPods: any[]
  fromExternal: boolean
  step1: boolean
  step2: boolean
  env: Env
  kindAction: [Kind | '', string]
  spec: any
  basic: any
} = {
  namespaces: [],
  labels: {},
  annotations: {},
  pods: [],
  networkTargetPods: [],
  // New Experiment needed
  fromExternal: false,
  step1: false,
  step2: false,
  env: 'k8s',
  kindAction: ['', ''],
  spec: {},
  basic: {},
}

const experimentsSlice = createSlice({
  name: 'experiments',
  initialState,
  reducers: {
    clearPods(state) {
      state.pods = []
      state.networkTargetPods = []
    },
    clearNetworkTargetPods(state) {
      state.networkTargetPods = []
    },
    setStep1(state, action: PayloadAction<boolean>) {
      state.step1 = action.payload
    },
    setStep2(state, action: PayloadAction<boolean>) {
      state.step2 = action.payload
    },
    setEnv(state, action: PayloadAction<Env>) {
      state.env = action.payload
    },
    setKindAction(state, action) {
      state.kindAction = action.payload
      state.spec = {}
    },
    setSpec(state, action) {
      state.spec = action.payload
    },
    setBasic(state, action) {
      state.basic = action.payload
    },
    setExternalExperiment(state, action: PayloadAction<any>) {
      const { kindAction, spec, basic } = action.payload

      state.fromExternal = true
      state.kindAction = kindAction
      state.spec = spec
      state.basic = basic
    },
    resetNewExperiment(state) {
      state.pods = []
      state.networkTargetPods = []
      state.fromExternal = false
      state.step1 = false
      state.step2 = false
      state.kindAction = ['', '']
      state.spec = {}
      state.basic = {}
    },
  },
  extraReducers: (builder) => {
    builder.addCase(getNamespaces.fulfilled, (state, action) => {
      state.namespaces = action.payload
    })
    builder.addCase(getLabels.fulfilled, (state, action) => {
      state.labels = action.payload
    })
    builder.addCase(getAnnotations.fulfilled, (state, action) => {
      state.annotations = action.payload
    })
    builder.addCase(getCommonPods.fulfilled, (state, action) => {
      state.pods = action.payload
    })
    builder.addCase(getNetworkTargetPods.fulfilled, (state, action) => {
      state.networkTargetPods = action.payload
    })
  },
})

export const {
  clearPods,
  clearNetworkTargetPods,
  setStep1,
  setStep2,
  setEnv,
  setKindAction,
  setSpec,
  setBasic,
  setExternalExperiment,
  resetNewExperiment,
} = experimentsSlice.actions

export default experimentsSlice.reducer
