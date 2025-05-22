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
import { Box, Divider, Typography } from '@mui/material'
import type { BoxProps } from '@mui/material'

import Space from '../Space'

interface PaperTopProps {
  title: React.ReactNode
  subtitle?: React.ReactNode
  h1?: boolean
  divider?: boolean
  boxProps?: BoxProps
}

const PaperTop: React.FC<PaperTopProps> = ({ title, subtitle, h1, divider, boxProps, children }) => (
  <Space width="100%">
    <Box {...boxProps} display="flex" justifyContent="space-between" alignItems="center">
      <Box flex={1}>
        <Typography variant={h1 ? 'h5' : 'h6'} component={h1 ? 'h1' : 'div'} fontWeight={h1 ? 'bold' : undefined}>
          {title}
        </Typography>
        {subtitle && (
          <Typography variant="body2" color="textSecondary">
            {subtitle}
          </Typography>
        )}
      </Box>
      {children}
    </Box>
    {divider && <Divider />}
  </Space>
)

export default PaperTop
