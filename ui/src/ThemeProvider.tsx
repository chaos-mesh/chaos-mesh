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
