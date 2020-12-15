import { Tooltip as MUITooltip } from '@material-ui/core'
import { withStyles } from '@material-ui/core/styles'

const Tooltip = withStyles((theme) => ({
  tooltip: {
    padding: theme.spacing(3),
    backgroundColor: theme.palette.common.black,
    color: theme.palette.common.white,
  },
  arrow: {
    color: theme.palette.common.black,
  },
}))(MUITooltip)

export default Tooltip
