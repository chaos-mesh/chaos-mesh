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
import LS from 'lib/localStorage'
import { createSlice } from '@reduxjs/toolkit'

export type Theme = 'light' | 'dark'

const initialState = {
  theme: (LS.get('theme') || 'light') as Theme,
  lang: LS.get('lang') || 'en',
  debugMode: LS.get('debug-mode') === 'true',
  enableKubeSystemNS: LS.get('enable-kube-system-ns') === 'true',
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
  },
})

export const { setTheme, setLang, setDebugMode, setEnableKubeSystemNS } = settingsSlice.actions

export default settingsSlice.reducer
