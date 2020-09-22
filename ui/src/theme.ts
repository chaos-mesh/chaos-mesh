import { createMuiTheme, responsiveFontSizes } from '@material-ui/core/styles'

// Design system
// https://material-ui.com/system/basics/
// The Default theme
// https://material-ui.com/customization/default-theme/
// How to customize
// https://material-ui.com/customization/theming/
const theme = responsiveFontSizes(
  createMuiTheme({
    palette: {
      primary: {
        main: '#172d72',
      },
    },
    spacing: (factor) => `${0.25 * factor}rem`,
  })
)

export const darkTheme = responsiveFontSizes(
  createMuiTheme({
    palette: {
      type: 'dark',
      primary: {
        main: '#9db0eb',
      },
    },
    spacing: (factor) => `${0.25 * factor}rem`,
  })
)

export default theme
