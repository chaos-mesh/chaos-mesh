import { createAsyncThunk, createSlice } from '@reduxjs/toolkit'

import { GlobalStatusAction } from './globalStatus.type'
import { StateOfExperiments } from 'api/experiments.type'
import api from 'api'

const defaultExperiments: StateOfExperiments = {
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

const globalStatusSlice = createSlice({
  name: 'globalStatus',
  initialState: {
    stateOfExperiments: defaultExperiments,
    needToRefreshExperiments: false,
  },
  reducers: {
    toggleNeedToRefreshExperiments(state) {
      state.needToRefreshExperiments = !state.needToRefreshExperiments
    },
  },
  extraReducers: (builder) => {
    builder.addCase(getStateofExperiments.fulfilled, (state, action: GlobalStatusAction) => {
      state.stateOfExperiments = action.payload
    })
  },
})

export const { toggleNeedToRefreshExperiments } = globalStatusSlice.actions

export default globalStatusSlice.reducer
