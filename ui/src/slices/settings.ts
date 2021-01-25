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
