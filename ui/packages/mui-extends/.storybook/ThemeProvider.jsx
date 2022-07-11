// This file is synchronized from '@ui/app/src/theme.ts'.
// All options are based on https://www.figma.com/file/2J6PVAaitQPQFHBtF5LbII/UI-Interface.
import { ThemeProvider, createTheme, responsiveFontSizes } from '@mui/material/styles'
import React from 'react'

const common = {
  spacing: 4,
  typography: {
    button: {
      textTransform: 'none',
    },
  },
}

// The light theme
const theme = responsiveFontSizes(
  createTheme({
    ...common,
    palette: {
      primary: {
        main: '#4159A9',
      },
      secondary: {
        main: '#595D71',
      },
      secondaryContainer: {
        main: '#DEE1F9',
      },
      onSecondaryContainer: {
        main: '#161B2C',
      },
      surfaceVariant: {
        main: '#E2E1EC',
      },
      onSurfaceVariant: {
        main: '#45464E',
      },
      background: {
        default: '#fafafa',
      },
      outline: {
        main: '#76767F',
      },
    },
  })
)

export default function ({ children }) {
  return <ThemeProvider theme={theme}>{children}</ThemeProvider>
}
