/*
 * Copyright 2022 Chaos Mesh Authors.
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
import { Theme } from '@mui/material/styles'

// TODO: remove this declaration when @mui/styles isn't used anymore. See: https://mui.com/guides/migration-v4/.
declare module '@mui/styles' {
  interface DefaultTheme extends Theme {}
}

declare module '@mui/material/styles' {
  interface Palette {
    secondaryContainer: Palette['primary']
    onSecondaryContainer: Palette['primary']
    surfaceVariant: Palette['primary']
    onSurfaceVariant: Palette['primary']
    outline: Palette['primary']
  }

  interface PaletteOptions {
    secondaryContainer: PaletteOptions['primary']
    onSecondaryContainer: PaletteOptions['primary']
    surfaceVariant: PaletteOptions['primary']
    onSurfaceVariant: PaletteOptions['primary']
    outline: PaletteOptions['primary']
  }
}
