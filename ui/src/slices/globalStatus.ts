import { Alert } from './globalStatus.type'
import { createSlice } from '@reduxjs/toolkit'

const initialState: {
  alert: Alert
  alertOpen: boolean
  searchModalOpen: boolean
  tokenIntercepterNumber: number
} = {
  alert: {
    type: 'success',
    message: '',
  },
  alertOpen: false,
  searchModalOpen: false,
  tokenIntercepterNumber: -1,
}

const globalStatusSlice = createSlice({
  name: 'globalStatus',
  initialState,
  reducers: {
    setAlert(state, action) {
      state.alert = action.payload
    },
    setAlertOpen(state, action) {
      state.alertOpen = action.payload
    },
    setSearchModalOpen(state, action) {
      state.searchModalOpen = action.payload
    },
    setTokenIntercepterNumber(state, action) {
      state.tokenIntercepterNumber = action.payload
    },
  },
})

export const { setAlert, setAlertOpen, setSearchModalOpen, setTokenIntercepterNumber } = globalStatusSlice.actions

export default globalStatusSlice.reducer
