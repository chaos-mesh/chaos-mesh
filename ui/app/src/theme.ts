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
// All options are based on https://www.figma.com/file/2J6PVAaitQPQFHBtF5LbII/UI-Interface.
import { ThemeOptions, createTheme, responsiveFontSizes } from '@mui/material/styles'

const common: ThemeOptions = {
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

export const darkTheme = responsiveFontSizes(
  createTheme({
    ...common,
    palette: {
      mode: 'dark',
      primary: {
        main: '#B4C4FF',
      },
      secondary: {
        main: '#C1C5DC',
      },
      secondaryContainer: {
        main: '#424659',
      },
      onSecondaryContainer: {
        main: '#DEE1F9',
      },
      surfaceVariant: {
        main: '#45464E',
      },
      onSurfaceVariant: {
        main: '#C6C6D0',
      },
      background: {
        paper: '#000',
        default: '#1B1B1F',
      },
      outline: {
        main: '#90909A',
      },
    },
  })
)

export default theme
