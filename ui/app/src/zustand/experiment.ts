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

import { Kind } from '@/components/NewExperimentNext/data/types'

export type Env = 'k8s' | 'physic'
export type KindAction = [Kind | '', string]

export interface AnyObject {
  [key: string]: AnyObject
}

export const useExperimentStore = create(
  combine(
    {
      fromExternal: false,
      step1: false,
      step2: false,
      env: 'k8s' as Env,
      kindAction: ['', ''] as KindAction,
      spec: {} as AnyObject,
      basic: {} as AnyObject,
    },
    (set) => ({
      actions: {
        setStep1: (step1: boolean) => {
          set({ step1 })
        },
        setStep2: (step2: boolean) => {
          set({ step2 })
        },
        setEnv: (env: Env) => {
          set({ env })
        },
        setKindAction: (kindAction: KindAction) => {
          set({ kindAction })
        },
        setSpec: (spec: AnyObject) => {
          set({ spec })
        },
        setBasic: (basic: AnyObject) => {
          set({ basic })
        },
        setExternalExp: ({
          env,
          kindAction,
          spec,
          basic,
        }: {
          env: Env
          kindAction: KindAction
          spec: AnyObject
          basic: AnyObject
        }) => {
          set({
            fromExternal: true,
            env,
            kindAction,
            spec,
            basic,
          })
        },
        reset: () => {
          set({
            fromExternal: false,
            step1: false,
            step2: false,
            kindAction: ['', ''],
            spec: {},
            basic: {},
          })
        },
      },
    }),
  ),
)

export const useExperimentActions = () => useExperimentStore((state) => state.actions)
