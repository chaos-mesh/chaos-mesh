import { createSlice } from '@reduxjs/toolkit'

export type Theme = 'light' | 'dark'

const initialState = {
  theme: (window.localStorage.getItem('chaos-mesh-theme') || 'light') as Theme,
  lang: window.localStorage.getItem('chaos-mesh-lang') || 'en',
}

const settingsSlice = createSlice({
  name: 'settings',
  initialState,
  reducers: {
    setTheme(state, action) {
      state.theme = action.payload

      window.localStorage.setItem('chaos-mesh-theme', action.payload)
    },
    setLang(state, action) {
      state.lang = action.payload

      window.localStorage.setItem('chaos-mesh-lang', action.payload)
    },
  },
})

export const { setTheme, setLang } = settingsSlice.actions

export default settingsSlice.reducer
