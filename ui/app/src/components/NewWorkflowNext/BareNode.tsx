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
import { Box, Button } from '@mui/material'
import type { ButtonProps } from '@mui/material'
import { forwardRef } from 'react'

import { iconByKind } from 'lib/byKind'

export type BareNodeProps = ButtonProps & {
  kind: string
}

export default forwardRef<HTMLSpanElement, BareNodeProps>(({ kind, sx, children, name, ...rest }, ref) => (
  <Button
    ref={ref}
    component="span"
    variant="outlined"
    color="secondary"
    size="small"
    startIcon={iconByKind(kind)}
    disableFocusRipple
    sx={{ alignItems: 'center', width: 200, ...sx }}
    children={<Box flex={1}>{children}</Box>}
    title={name}
    {...rest}
  />
))
