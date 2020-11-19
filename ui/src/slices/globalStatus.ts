import { PayloadAction, createSlice } from '@reduxjs/toolkit'

import { TokenFormValues } from 'components/Token'

export interface Alert {
  type: 'success' | 'warning' | 'error'
  message: string
}

const initialState: {
  alert: Alert
  alertOpen: boolean
  searchModalOpen: boolean
  tokenInterceptor: number
  tokens: TokenFormValues[]
  tokenName: string
  namespace: string
} = {
  alert: {
    type: 'success',
    message: '',
  },
  alertOpen: false,
  searchModalOpen: false,
  tokens: [],
  tokenInterceptor: -1,
  tokenName: '',
  namespace: 'default',
}

const globalStatusSlice = createSlice({
  name: 'globalStatus',
  initialState,
  reducers: {
    setAlert(state, action: PayloadAction<Alert>) {
      state.alert = action.payload
    },
    setAlertOpen(state, action: PayloadAction<boolean>) {
      state.alertOpen = action.payload
    },
    setSearchModalOpen(state, action: PayloadAction<boolean>) {
      state.searchModalOpen = action.payload
    },
    setTokens(state, action: PayloadAction<TokenFormValues[]>) {
      state.tokens = action.payload
    },
    setTokenInterceptor(state, action: PayloadAction<number>) {
      state.tokenInterceptor = action.payload
    },
    setTokenName(state, action: PayloadAction<string>) {
      state.tokenName = action.payload
    },
    setNameSpace(state, action: PayloadAction<string>) {
      state.namespace = action.payload
    },
  },
})

export const {
  setAlert,
  setAlertOpen,
  setSearchModalOpen,
  setTokens,
  setTokenInterceptor,
  setTokenName,
  setNameSpace,
} = globalStatusSlice.actions

export default globalStatusSlice.reducer
