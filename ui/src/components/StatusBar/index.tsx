import { Box, Button, Paper, Toolbar } from '@material-ui/core'
import React, { useEffect } from 'react'
import { RootState, useStoreDispatch } from 'store'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import NewExperiment from 'components/NewExperiment'
import { StateOfExperiments } from 'slices/globalStatus.type'
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
        flexDirection: 'column-reverse',
        alignItems: 'start',
        paddingTop: sp3,
        paddingBottom: sp3,
      },
    },
    new: {
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
  state: StateOfExperiments
}

export const CurrentStatus: React.FC<CurrentStatusProps> = ({ state }) => {
  const classes = useStyles()

  return (
    <Box display="flex">
      <Button className={classes.statusButton} variant="outlined">
        Total: {state.total}
      </Button>
      <Button className={classes.statusButton} variant="outlined" color="primary">
        Running: {state.running}
      </Button>
      <Button className={classes.statusButton} variant="outlined" color="secondary">
        Failed: {state.failed}
      </Button>
      <Button className={classes.statusButton} variant="outlined">
        Finished: {state.finished}
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
        <Box className={classes.new}>
          <NewExperiment />
        </Box>

        <CurrentStatus state={stateOfExperiments} />
      </Toolbar>
    </Paper>
  )
}

export default StatusBar
