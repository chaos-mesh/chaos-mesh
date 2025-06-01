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

export type SystemTheme = 'light' | 'dark'

export const useSystemStore = create(
  combine({ theme: (LS.get('theme') || 'light') as SystemTheme, lang: LS.get('lang') || 'en' }, (set) => ({
    actions: {
      setTheme: (theme: SystemTheme) => set({ theme }),
      setLang: (lang: string) => set({ lang }),
    },
  })),
)

export const useSystemActions = () => useSystemStore((state) => state.actions)
