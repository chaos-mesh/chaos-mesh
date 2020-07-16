import { Box, Button, Paper, Toolbar } from '@material-ui/core'
import React, { useEffect } from 'react'
import { RootState, useStoreDispatch } from 'store'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import NewExperiment from 'components/NewExperiment'
import { StateOfExperiments } from 'api/experiments.type'
import clsx from 'clsx'
import { getStateofExperiments } from 'slices/experiments'
import useErrorButtonStyles from 'lib/styles/errorButton'
import { useSelector } from 'react-redux'

const useStyles = makeStyles((theme: Theme) => {
  const sp3 = theme.spacing(3)

  return createStyles({
    root: {
      paddingLeft: sp3,
      paddingRight: sp3,
      [theme.breakpoints.down('xs')]: {
        height: 0,
      },
    },
    toolbar: {
      ...theme.mixins.toolbar,
      justifyContent: 'space-between',
      [theme.breakpoints.down(700)]: {
        flexDirection: 'column',
        alignItems: 'start',
        '& > *': {
          marginTop: sp3,
          '&:last-child': {
            marginBottom: sp3,
          },
        },
      },
    },
    currentStatus: {
      display: 'flex',
      [theme.breakpoints.down('xs')]: {
        display: 'none',
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
  const errorButton = useErrorButtonStyles()

  const data = [
    {
      label: 'Running',
      value: state.running,
      color: 'primary' as 'primary',
    },
    {
      label: 'Paused',
      value: state.paused,
    },
    {
      label: 'Failed',
      value: state.failed,
      className: errorButton.root,
    },
  ]

  return (
    <Box className={classes.currentStatus}>
      {data.map((d) => (
        <Button
          key={d.label}
          className={clsx(classes.statusButton, d.className)}
          variant="outlined"
          size="small"
          color={d.color}
        >
          {d.label}: {d.value}
        </Button>
      ))}
    </Box>
  )
}

const StatusBar = () => {
  const classes = useStyles()

  const stateOfExperiments = useSelector((state: RootState) => state.experiments.stateOfExperiments)
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

        <CurrentStatus state={stateOfExperiments} />
      </Toolbar>
    </Paper>
  )
}

export default StatusBar
