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
import { createSlice } from '@reduxjs/toolkit'

import LS from 'lib/localStorage'

export type Theme = 'light' | 'dark'

const initialState = {
  theme: (LS.get('theme') || 'light') as Theme,
  lang: LS.get('lang') || 'en',
  debugMode: LS.get('debug-mode') === 'true',
  enableKubeSystemNS: LS.get('enable-kube-system-ns') === 'true',
  useNewPhysicalMachine: LS.get('use-new-physical-machine') === 'true',
  useNextWorkflowInterface: (LS.get('use-next-workflow-interface') || 'true') === 'true',
}

const settingsSlice = createSlice({
  name: 'settings',
  initialState,
  reducers: {
    setTheme(state, action) {
      state.theme = action.payload

      LS.set('theme', action.payload)
    },
    setLang(state, action) {
      state.lang = action.payload

      LS.set('lang', action.payload)
    },
    setDebugMode(state, action) {
      state.debugMode = action.payload

      LS.set('debug-mode', action.payload)
    },
    setEnableKubeSystemNS(state, action) {
      state.enableKubeSystemNS = action.payload

      LS.set('enable-kube-system-ns', action.payload)
    },
    setUseNewPhysicalMachine(state, action) {
      state.useNewPhysicalMachine = action.payload

      LS.set('use-new-physical-machine', action.payload)
    },
    setUseNextWorkflowInterface(state, action) {
      state.useNextWorkflowInterface = action.payload

      LS.set('use-next-workflow-interface', action.payload)
    },
  },
})

export const {
  setTheme,
  setLang,
  setDebugMode,
  setEnableKubeSystemNS,
  setUseNewPhysicalMachine,
  setUseNextWorkflowInterface,
} = settingsSlice.actions

export default settingsSlice.reducer
