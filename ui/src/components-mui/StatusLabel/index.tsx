import CheckCircleIcon from '@material-ui/icons/CheckCircle'
import { Chip } from '@material-ui/core'
import { Experiment } from 'api/experiments.type'
import PauseCircleFilledIcon from '@material-ui/icons/PauseCircleFilled'
import T from 'components/T'
import { Workflow } from 'api/workflows.type'
import { useIntl } from 'react-intl'

interface StatusLabelProps {
  status: Workflow['status'] | Experiment['status']
}

const StatusLabel: React.FC<StatusLabelProps> = ({ status }) => {
  const intl = useIntl()

  let label
  switch (status) {
    case 'Succeed':
      label = T(`workflow.status.${status}`, intl)

      break
    case 'injecting':
    case 'running':
    case 'finished':
    case 'paused':
      label = T(`experiments.status.${status}`, intl)

      break
  }

  let color
  switch (status) {
    case 'injecting':
    case 'running':
      color = 'info.main'

      break
    case 'paused':
      color = 'grey.500'

      break
    case 'Succeed':
    case 'finished':
      color = 'success.main'

      break
  }

  let icon
  switch (status) {
    case 'Succeed':
    case 'finished':
      icon = <CheckCircleIcon sx={{ color }} />

      break
    case 'injecting':
    case 'running':
      break
    case 'paused':
      icon = <PauseCircleFilledIcon sx={{ color }} />

      break
  }

  return <Chip variant="outlined" size="small" icon={icon} label={label} sx={{ color, borderColor: color }} />
}

export default StatusLabel
