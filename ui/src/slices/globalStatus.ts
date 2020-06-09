import { Alert, GlobalStatusAction } from './globalStatus.type'
import { createAsyncThunk, createSlice } from '@reduxjs/toolkit'

import { StateOfExperiments } from 'api/experiments.type'
import api from 'api'

const defaultExperiments = {
  total: 0,
  running: 0,
  paused: 0,
  failed: 0,
  finished: 0,
}

export const getStateofExperiments = createAsyncThunk('experiments/state', async () => {
  const resp = await api.experiments.state()

  return resp.data
})

const initialState: {
  alert: Alert
  alertOpen: boolean
  stateOfExperiments: StateOfExperiments
  needToRefreshExperiments: boolean
} = {
  alert: {
    type: 'success',
    message: '',
  },
  alertOpen: false,
  stateOfExperiments: defaultExperiments,
  needToRefreshExperiments: false,
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
    setNeedToRefreshExperiments(state, action) {
      state.needToRefreshExperiments = action.payload
    },
  },
  extraReducers: (builder) => {
    builder.addCase(getStateofExperiments.fulfilled, (state, action: GlobalStatusAction) => {
      state.stateOfExperiments = action.payload
    })
  },
})

export const { setAlert, setAlertOpen, setNeedToRefreshExperiments } = globalStatusSlice.actions

export default globalStatusSlice.reducer
