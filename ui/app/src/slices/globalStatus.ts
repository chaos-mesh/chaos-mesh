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

import { TokenFormValues } from '@/components/Token'

import LS from '@/lib/localStorage'

const initialState: {
  authOpen: boolean
  namespace: string
  tokens: TokenFormValues[]
  tokenName: string
} = {
  authOpen: false,
  namespace: 'All',
  tokens: [],
  tokenName: '',
}

const globalStatusSlice = createSlice({
  name: 'globalStatus',
  initialState,
  reducers: {
    setAuthOpen(state, action: PayloadAction<boolean>) {
      state.authOpen = action.payload
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
    removeToken(state) {
      state.tokenName = ''
      state.tokens = []

      LS.remove('token')
      LS.remove('token-name')
    },
  },
})

export const { setAuthOpen, setNameSpace, setTokens, setTokenName, removeToken } = globalStatusSlice.actions

export default globalStatusSlice.reducer
