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
    },
  })
)

export default theme
