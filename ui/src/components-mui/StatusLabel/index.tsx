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
