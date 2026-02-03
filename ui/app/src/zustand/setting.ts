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

import LS from '@/lib/localStorage'

export const useSettingStore = create(
  combine(
    {
      debugMode: LS.get('debug-mode') === 'true',
      enableKubeSystemNS: LS.get('enable-kube-system-ns') === 'true',
      useNewPhysicalMachine: LS.get('use-new-physical-machine') === 'true',
      eventTimeFormat: (LS.get('event-time-format') || 'relative') as 'relative' | 'absolute',
    },
    (set) => ({
      actions: {
        setDebugMode: (debugMode: boolean) => {
          set({ debugMode })
          LS.set('debug-mode', debugMode ? 'true' : 'false')
        },
        setEnableKubeSystemNS: (enableKubeSystemNS: boolean) => {
          set({ enableKubeSystemNS })
          LS.set('enable-kube-system-ns', enableKubeSystemNS ? 'true' : 'false')
        },
        setUseNewPhysicalMachine: (useNewPhysicalMachine: boolean) => {
          set({ useNewPhysicalMachine })
          LS.set('use-new-physical-machine', useNewPhysicalMachine ? 'true' : 'false')
        },
        setEventTimeFormat: (eventTimeFormat: 'relative' | 'absolute') => {
          set({ eventTimeFormat })
          LS.set('event-time-format', eventTimeFormat)
        },
      },
    }),
  ),
)

export const useSettingActions = () => useSettingStore((state) => state.actions)
