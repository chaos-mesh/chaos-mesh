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
import React from 'react'
import { create } from 'zustand'
import { combine } from 'zustand/middleware'

export interface Alert {
  type: 'success' | 'warning' | 'error'
  message: React.ReactNode
}

export interface Confirm {
  title: string
  description?: React.ReactNode
  handle?: () => void
}

export const useComponentStore = create(
  combine(
    {
      alert: {
        type: 'success',
        message: '' as React.ReactNode,
      } as Alert,
      alertOpen: false,
      confirm: {
        title: '',
      } as Confirm,
      authOpen: false,
      confirmOpen: false,
    },
    (set) => ({
      actions: {
        setAlert: (alert: Alert) => set({ alert, alertOpen: true }),
        setAlertOpen: (open: boolean) => set({ alertOpen: open }),
        setConfirm: (confirm: Confirm) => set({ confirm, confirmOpen: true }),
        setConfirmOpen: (open: boolean) => set({ confirmOpen: open }),
      },
    }),
  ),
)

export const useComponentActions = () => useComponentStore((state) => state.actions)
