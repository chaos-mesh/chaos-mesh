import red from '@material-ui/core/colors/red'
import { createMuiTheme, responsiveFontSizes } from '@material-ui/core/styles'

// Material UI Style System:
// https://material-ui.com/system/basics/
// Customization API
// https://material-ui.com/customization/spacing/

// A custom theme for this app
let theme = createMuiTheme({
  palette: {
    primary: {
      main: '#172d72',
    },
    secondary: {
      main: '#880e4f',
    },
    error: {
      main: red.A400,
    },
    background: {
      default: '#fff',
    },
  },
  spacing: (factor: number) => `${0.25 * factor}rem`, // (Bootstrap strategy),
})

theme = responsiveFontSizes(theme)

export default theme
