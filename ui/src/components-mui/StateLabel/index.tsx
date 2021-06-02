import { Button, ButtonProps, CircularProgress } from '@material-ui/core'

import { StateOfExperimentsEnum } from 'api/experiments.type'
import clsx from 'clsx'
import { makeStyles } from '@material-ui/styles'

const useStyles = makeStyles((theme) => ({
  info: {
    color: theme.palette.info.main,
    borderColor: theme.palette.info.main,
  },
}))

interface Props {
  state: StateOfExperimentsEnum
}

function StateLabel(props: ButtonProps & Props) {
  const classes = useStyles()

  let icon
  let className

  switch (props.state) {
    case StateOfExperimentsEnum.Running:
      icon = <CircularProgress size={15} color="inherit" disableShrink />
      className = classes.info
  }

  return (
    <Button {...props} className={clsx(props.className, className)} variant="outlined" size="small" startIcon={icon}>
      {props.children}
    </Button>
  )
}

export default StateLabel
