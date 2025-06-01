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
    },
    (set) => ({
      actions: {
        setDebugMode: (debugMode: boolean) => set({ debugMode }),
        setEnableKubeSystemNS: (enableKubeSystemNS: boolean) => set({ enableKubeSystemNS }),
        setUseNewPhysicalMachine: (useNewPhysicalMachine: boolean) => set({ useNewPhysicalMachine }),
      },
    }),
  ),
)

export const useSettingActions = () => useSettingStore((state) => state.actions)
