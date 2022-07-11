/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import { PayloadAction, createSlice } from '@reduxjs/toolkit'
import { ConfigChaosDashboardConfig } from 'openapi'
import React from 'react'

import { TokenFormValues } from 'components/Token'

import LS from 'lib/localStorage'

export interface Alert {
  type: 'success' | 'warning' | 'error'
  message: React.ReactNode
}

export interface Confirm {
  title: string
  description?: React.ReactNode
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
  gcpSecurityMode: boolean
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
  gcpSecurityMode: false,
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
    setConfig(state, action: PayloadAction<ConfigChaosDashboardConfig>) {
      state.securityMode = action.payload.security_mode!
      state.dnsServerCreate = action.payload.dns_server_create!
      state.gcpSecurityMode = action.payload.gcp_security_mode!
      state.version = action.payload.version!
    },
    setNameSpace(state, action: PayloadAction<string>) {
      const ns = action.payload

      state.namespace = ns

      LS.set('global-namespace', ns)
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

export const { setAlert, setAlertOpen, setConfirm, setConfirmOpen, setConfig, setNameSpace, setTokens, setTokenName } =
  globalStatusSlice.actions

export default globalStatusSlice.reducer
