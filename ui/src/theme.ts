import { blue, green, orange, red } from '@material-ui/core/colors'
import { createMuiTheme, responsiveFontSizes } from '@material-ui/core/styles'

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
      secondary: {
        main: '#72172d',
      },
      error: {
        main: red.A400,
      },
      warning: {
        main: orange.A400,
      },
      info: {
        main: blue.A400,
      },
      success: {
        main: green.A400,
      },
      background: {
        default: '#f5f5f5',
      },
    },
    spacing: (factor) => `${0.25 * factor}rem`, // (Bootstrap strategy)
  })
)

export default theme
