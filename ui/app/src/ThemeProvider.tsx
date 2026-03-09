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
import { ThemeProvider as MuiThemeProvider } from '@mui/material/styles'
import { useMemo } from 'react'

import theme, { darkTheme } from './theme'
import { useSystemStore } from './zustand/system'

const ThemeProvider: ReactFCWithChildren = ({ children }) => {
  const t = useSystemStore((state) => state.theme)
  const globalTheme = useMemo(() => (t === 'light' ? theme : darkTheme), [t])

  return <MuiThemeProvider theme={globalTheme}>{children}</MuiThemeProvider>
}

export default ThemeProvider
