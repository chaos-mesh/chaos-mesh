import { PayloadAction, createSlice } from '@reduxjs/toolkit'

import LS from 'lib/localStorage'
import { TokenFormValues } from 'components/Token'

export interface Alert {
  type: 'success' | 'warning' | 'error'
  message: string
}

const initialState: {
  alert: Alert
  alertOpen: boolean
  searchModalOpen: boolean
  namespace: string
  securityMode: boolean
  tokens: TokenFormValues[]
  tokenName: string
} = {
  alert: {
    type: 'success',
    message: '',
  },
  alertOpen: false,
  searchModalOpen: false,
  namespace: 'All',
  securityMode: true,
  tokens: [],
  tokenName: '',
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
    setNameSpace(state, action: PayloadAction<string>) {
      const ns = action.payload

      state.namespace = ns

      LS.set('global-namespace', ns)
    },
    setSecurityMode(state, action: PayloadAction<boolean>) {
      state.securityMode = action.payload
    },
    setTokens(state, action: PayloadAction<TokenFormValues[]>) {
      const tokens = action.payload

      state.tokens = tokens

      LS.set('token', JSON.stringify(tokens))
    },
    setTokenName(state, action: PayloadAction<string>) {
      const name = action.payload

      state.tokenName = name

      LS.set('token-name', name)
    },
  },
})

export const {
  setAlert,
  setAlertOpen,
  setSearchModalOpen,
  setNameSpace,
  setSecurityMode,
  setTokens,
  setTokenName,
} = globalStatusSlice.actions

export default globalStatusSlice.reducer
