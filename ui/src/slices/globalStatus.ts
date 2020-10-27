import { Alert } from './globalStatus.type'
import { createSlice } from '@reduxjs/toolkit'

const initialState: {
  alert: Alert
  alertOpen: boolean
  searchModalOpen: boolean
} = {
  alert: {
    type: 'success',
    message: '',
  },
  alertOpen: false,
  searchModalOpen: false,
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
  },
})

export const { setAlert, setAlertOpen, setSearchModalOpen } = globalStatusSlice.actions

export default globalStatusSlice.reducer
