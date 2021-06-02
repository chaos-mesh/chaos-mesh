import { Tooltip as MUITooltip } from '@material-ui/core'
import { withStyles } from '@material-ui/styles'

const Tooltip = withStyles((theme) => ({
  tooltip: {
    padding: theme.spacing(3),
    backgroundColor: theme.palette.background.default,
    color: theme.palette.text.primary,
    border: `1px solid ${theme.palette.divider}`,
  },
}))(MUITooltip)

export default Tooltip
