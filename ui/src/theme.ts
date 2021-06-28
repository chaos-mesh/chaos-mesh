import { ThemeOptions, createTheme, responsiveFontSizes } from '@material-ui/core/styles'

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
    secondary: {
      main: '#d32f2f',
    },
    background: {
      default: '#fafafa',
    },
  },
  spacing: (factor: number) => `${0.25 * factor}rem`,
}

const theme = responsiveFontSizes(createTheme(common))

export const darkTheme = responsiveFontSizes(
  createTheme({
    ...common,
    palette: {
      mode: 'dark',
      primary: {
        main: '#9db0eb',
      },
      secondary: {
        main: '#f44336',
      },
      background: {
        paper: '#424242',
        default: '#303030',
      },
    },
  })
)

export default theme
