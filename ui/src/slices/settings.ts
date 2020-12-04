import LS from 'lib/localStorage'
import { createSlice } from '@reduxjs/toolkit'

export type Theme = 'light' | 'dark'

const initialState = {
  theme: (LS.get('theme') || 'light') as Theme,
  lang: LS.get('lang') || 'en',
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
  },
})

export const { setTheme, setLang } = settingsSlice.actions

export default settingsSlice.reducer
