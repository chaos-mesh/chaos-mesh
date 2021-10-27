/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
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
