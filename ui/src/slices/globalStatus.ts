import { Alert } from './globalStatus.type'
import { createSlice } from '@reduxjs/toolkit'

const initialState: {
  alert: Alert
  alertOpen: boolean
} = {
  alert: {
    type: 'success',
    message: '',
  },
  alertOpen: false,
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
  },
})

export const { setAlert, setAlertOpen } = globalStatusSlice.actions

export default globalStatusSlice.reducer
