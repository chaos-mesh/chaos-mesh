import { createSlice } from '@reduxjs/toolkit'

const initialState = {
  lang: window.localStorage.getItem('chaos-mesh-lang') || 'en',
}

const settingsSlice = createSlice({
  name: 'settings',
  initialState,
  reducers: {
    setLang(state, action) {
      state.lang = action.payload

      window.localStorage.setItem('chaos-mesh-lang', action.payload)
    },
  },
})

export const { setLang } = settingsSlice.actions

export default settingsSlice.reducer
