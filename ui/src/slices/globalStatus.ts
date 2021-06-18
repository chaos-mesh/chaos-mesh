import { PayloadAction, createSlice } from '@reduxjs/toolkit'

import { Config } from 'api/common.type'
import LS from 'lib/localStorage'
import { TokenFormValues } from 'components/Token'

export interface Alert {
  type: 'success' | 'warning' | 'error'
  message: string
}

export interface Confirm {
  title: string
  description?: string
  handle?: () => void
  [key: string]: any
}

const initialState: {
  alert: Alert
  alertOpen: boolean
  confirm: Confirm
  confirmOpen: boolean // control global confirm dialog
  namespace: string
  securityMode: boolean
  dnsServerCreate: boolean
  version: string
  tokens: TokenFormValues[]
  tokenName: string
} = {
  alert: {
    type: 'success',
    message: '',
  },
  alertOpen: false,
  confirm: {
    title: '',
    description: '',
  },
  confirmOpen: false,
  namespace: 'All',
  securityMode: true,
  dnsServerCreate: false,
  version: '',
  tokens: [],
  tokenName: '',
}

const globalStatusSlice = createSlice({
  name: 'globalStatus',
  initialState,
  reducers: {
    setAlert(state, action: PayloadAction<Alert>) {
      state.alert = action.payload
      state.alertOpen = true
    },
    setAlertOpen(state, action: PayloadAction<boolean>) {
      state.alertOpen = action.payload
    },
    setConfirm(state, action: PayloadAction<Confirm>) {
      state.confirm = action.payload
      state.confirmOpen = true
    },
    setConfirmOpen(state, action: PayloadAction<boolean>) {
      state.confirmOpen = action.payload
    },
    setNameSpace(state, action: PayloadAction<string>) {
      const ns = action.payload

      state.namespace = ns

      LS.set('global-namespace', ns)
    },
    setConfig(state, action: PayloadAction<Config>) {
      state.securityMode = action.payload.security_mode
      state.dnsServerCreate = action.payload.dns_server_create
      state.version = action.payload.version
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

export const { setAlert, setAlertOpen, setConfirm, setConfirmOpen, setNameSpace, setConfig, setTokens, setTokenName } =
  globalStatusSlice.actions

export default globalStatusSlice.reducer
