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

import { Kind } from 'components/NewExperimentNext/data/types'

export type Env = 'k8s' | 'physic'

const initialState: {
  fromExternal: boolean
  step1: boolean
  step2: boolean
  env: Env
  kindAction: [Kind | '', string]
  spec: any
  basic: any
} = {
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
      state.fromExternal = false
      state.step1 = false
      state.step2 = false
      state.kindAction = ['', '']
      state.spec = {}
      state.basic = {}
    },
  },
})

export const {
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
