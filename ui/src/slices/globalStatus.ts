import { Alert } from './globalStatus.type'
import { IntlShape } from 'react-intl'
import { createSlice } from '@reduxjs/toolkit'

const initialState: {
  alert: Alert
  alertOpen: boolean
  intl?: IntlShape
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
    setIntl(state, action) {
      state.intl = action.payload
    },
  },
})

export const { setAlert, setAlertOpen, setIntl } = globalStatusSlice.actions

export default globalStatusSlice.reducer
