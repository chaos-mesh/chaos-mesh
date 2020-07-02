import { createMuiTheme, responsiveFontSizes } from '@material-ui/core/styles'
import { green, pink } from '@material-ui/core/colors'

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
      secondary: {
        main: pink[900],
      },
      success: {
        main: green[700],
      },
    },
    spacing: (factor) => `${0.25 * factor}rem`,
  })
)

export default theme
