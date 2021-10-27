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
import { Chip, CircularProgress, useTheme } from '@material-ui/core'

import CheckCircleIcon from '@material-ui/icons/CheckCircle'
import ErrorIcon from '@material-ui/icons/Error'
import { Experiment } from 'api/experiments.type'
import HelpIcon from '@material-ui/icons/Help'
import PauseCircleFilledIcon from '@material-ui/icons/PauseCircleFilled'
import T from 'components/T'
import { Workflow } from 'api/workflows.type'
import { useIntl } from 'react-intl'

interface StatusLabelProps {
  status: Workflow['status'] | Experiment['status']
}

const StatusLabel: React.FC<StatusLabelProps> = ({ status }) => {
  const intl = useIntl()
  const theme = useTheme()

  const label = T(`status.${status}`, intl)

  let color
  switch (status) {
    case 'injecting':
    case 'running':
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
    case 'finished':
      icon = <CheckCircleIcon style={{ color }} />

      break
    case 'injecting':
    case 'running':
      icon = (
        <CircularProgress
          size={15}
          disableShrink
          sx={{ ml: (theme) => `${theme.spacing(1)} !important`, color: `${color} !important` }}
        />
      )

      break
    case 'paused':
      icon = <PauseCircleFilledIcon style={{ color }} />

      break
    case 'unknown':
      icon = <HelpIcon style={{ color }} />

      break
    case 'failed':
      icon = <ErrorIcon style={{ color }} />

      break
  }

  return <Chip variant="outlined" size="small" icon={icon} label={label} sx={{ color, borderColor: color }} />
}

export default StatusLabel
