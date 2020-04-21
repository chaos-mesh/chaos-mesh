import { red, green, blue } from '@material-ui/core/colors'
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
    success: {
      main: green.A400,
    },
    info: {
      main: blue.A400,
    },
    background: {
      default: '#f5f5f5',
      paper: '#fff',
    },
  },
  spacing: (factor: number) => `${0.25 * factor}rem`, // (Bootstrap strategy),
})

theme = responsiveFontSizes(theme)

export default theme
