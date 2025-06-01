/*
 * Copyright 2025 Chaos Mesh Authors.
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
import { create } from 'zustand'
import { combine } from 'zustand/middleware'

import type { TokenFormValues } from '@/components/Token'

import LS from '@/lib/localStorage'

export const useAuthStore = create(
  combine(
    {
      authOpen: false,
      namespace: 'All',
      tokens: [] as TokenFormValues[],
      tokenName: '',
    },
    (set) => ({
      actions: {
        setAuthOpen: (authOpen: boolean) => set({ authOpen }),
        setNameSpace: (namespace: string) => {
          set({ namespace })
          LS.set('global-namespace', namespace)
        },
        setTokens: (tokens: TokenFormValues[]) => {
          set({ tokens })
          LS.set('token', JSON.stringify(tokens))
        },
        setTokenName: (tokenName: string) => {
          set({ tokenName })
          LS.set('token-name', tokenName)
        },
        removeToken: () => {
          set({ authOpen: true, tokens: [], tokenName: '' })
          LS.remove('token')
          LS.remove('token-name')
        },
      },
    }),
  ),
)

export const useAuthActions = () => useAuthStore((state) => state.actions)
