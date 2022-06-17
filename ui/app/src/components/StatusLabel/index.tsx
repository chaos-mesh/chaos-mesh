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
import { CircularProgress, Typography, styled, useTheme } from '@mui/material'

import Space from '@ui/mui-extends/esm/Space'

import { T } from 'components/T'

const Circle = styled('div')((props) => ({
  width: 8,
  height: 8,
  backgroundColor: props.color,
  borderRadius: '50%',
}))

interface StatusLabelProps {
  status: string
}

const StatusLabel: React.FC<StatusLabelProps> = ({ status }) => {
  const theme = useTheme()

  const label = <T id={`status.${status}`} />

  let color
  switch (status) {
    case 'injecting':
    case 'running':
    case 'deleting':
      color = theme.palette.primary.main

      break
    case 'paused':
    case 'unknown':
      color = theme.palette.grey[500]

      break
    case 'finished':
      color = theme.palette.success.main

      break
    case 'failed':
      color = theme.palette.error.main

      break
  }

  let icon
  switch (status) {
    case 'injecting':
    case 'running':
    case 'deleting':
      icon = <CircularProgress size={12} disableShrink sx={{ color }} />

      break
  }

  return (
    <Space spacing={1} direction="row" alignItems="center">
      {icon || <Circle color={color} />}
      <Typography variant="body2" fontWeight="500" sx={{ color }}>
        {label}
      </Typography>
    </Space>
  )
}

export default StatusLabel
