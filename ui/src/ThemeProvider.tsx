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
import { ThemeProvider as MUIThemeProvider, StyledEngineProvider } from '@material-ui/core/styles'
import customTheme, { darkTheme as customDarkTheme } from 'theme'

import { useMemo } from 'react'
import { useStoreSelector } from 'store'

const ThemeProvider: React.FC = ({ children }) => {
  const { theme } = useStoreSelector((state) => state.settings)
  const globalTheme = useMemo(() => (theme === 'light' ? customTheme : customDarkTheme), [theme])

  return (
    <MUIThemeProvider theme={globalTheme}>
      <StyledEngineProvider injectFirst>{children}</StyledEngineProvider>
    </MUIThemeProvider>
  )
}

export default ThemeProvider
