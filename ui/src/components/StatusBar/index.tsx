import { Box, Button, Paper, Toolbar, useMediaQuery, useTheme } from '@material-ui/core'
import React, { useEffect } from 'react'
import { RootState, useStoreDispatch } from 'store'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import NewExperiment from 'components/NewExperiment'
import { StateOfExperiments } from 'api/experiments.type'
import { getStateofExperiments } from 'slices/globalStatus'
import { useSelector } from 'react-redux'

const useStyles = makeStyles((theme: Theme) => {
  const sp3 = theme.spacing(3)

  return createStyles({
    root: {
      paddingLeft: sp3,
      paddingRight: sp3,
    },
    toolbar: {
      ...theme.mixins.toolbar,
      justifyContent: 'space-between',
      [theme.breakpoints.down('sm')]: {
        paddingBottom: sp3,
      },
    },
    currentStatus: {
      display: 'flex',
      [theme.breakpoints.down('sm')]: {
        marginTop: sp3,
      },
    },
    statusButton: {
      marginRight: sp3,
      '&:last-child': {
        marginRight: 0,
      },
    },
  })
})

interface CurrentStatusProps {
  classes: Record<'currentStatus' | 'statusButton', string>
  state: StateOfExperiments
}

export const CurrentStatus: React.FC<CurrentStatusProps> = ({ classes, state }) => {
  const theme = useTheme()
  const isMobileScreen = useMediaQuery(theme.breakpoints.down('sm'))

  const props = {
    className: classes.statusButton,
    variant: 'outlined' as 'outlined',
    size: isMobileScreen ? ('small' as 'small') : ('medium' as 'medium'),
  }

  return (
    <Box className={classes.currentStatus}>
      <Button {...props} color="primary">
        Running: {state.running}
      </Button>
      <Button {...props}>Finished: {state.finished}</Button>
      <Button {...props} color="secondary">
        Failed: {state.failed}
      </Button>
    </Box>
  )
}

const StatusBar = () => {
  const classes = useStyles()

  const stateOfExperiments = useSelector((state: RootState) => state.globalStatus.stateOfExperiments)
  const dispatch = useStoreDispatch()

  useEffect(() => {
    dispatch(getStateofExperiments())

    const id = setInterval(() => dispatch(getStateofExperiments()), 60000)

    return () => clearInterval(id)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <Paper className={classes.root} elevation={0}>
      <Toolbar className={classes.toolbar}>
        <NewExperiment />

        <CurrentStatus classes={classes} state={stateOfExperiments} />
      </Toolbar>
    </Paper>
  )
}

export default StatusBar
