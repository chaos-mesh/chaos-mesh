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
import { Box, Button } from '@mui/material'
import type { ButtonProps } from '@mui/material'

import { T } from 'components/T'

export default function Submit({ sx, ...rest }: ButtonProps) {
  return (
    <Box>
      <Button
        {...rest}
        type={rest.onClick ? undefined : 'submit'}
        variant="contained"
        size="small"
        fullWidth
        sx={{ mt: 3, ...sx }}
      >
        <T id="common.submit" />
      </Button>
    </Box>
  )
}
