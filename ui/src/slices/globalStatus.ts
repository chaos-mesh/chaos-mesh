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
  tokenName: '',
  namespace: 'All',
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
      const tokens = action.payload

      state.tokens = tokens

      LS.set('token', JSON.stringify(tokens))
    },
    setTokenName(state, action: PayloadAction<string>) {
      const name = action.payload

      state.tokenName = name

      LS.set('token-name', name)
    },
    setNameSpace(state, action: PayloadAction<string>) {
      const ns = action.payload

      state.namespace = ns

      LS.set('global-namespace', ns)
    },
  },
})

export const {
  setAlert,
  setAlertOpen,
  setSearchModalOpen,
  setTokens,
  setTokenName,
  setNameSpace,
} = globalStatusSlice.actions

export default globalStatusSlice.reducer
