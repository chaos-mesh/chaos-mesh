import { ThemeOptions, createMuiTheme, responsiveFontSizes } from '@material-ui/core/styles'

// Design system
// https://material-ui.com/system/basics/
// The Default theme
// https://material-ui.com/customization/default-theme/
// How to customize
// https://material-ui.com/customization/theming/
const common: ThemeOptions = {
  mixins: {
    toolbar: {
      minHeight: 56,
    },
  },
  palette: {
    primary: {
      main: '#172d72',
    },
  },
  spacing: (factor) => `${0.25 * factor}rem`,
}

const theme = responsiveFontSizes(createMuiTheme(common))

export const darkTheme = responsiveFontSizes(
  createMuiTheme({
    ...common,
    palette: {
      type: 'dark',
      primary: {
        main: '#9db0eb',
      },
    },
  })
)

export default theme
