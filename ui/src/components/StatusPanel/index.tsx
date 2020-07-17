import { Box, Grid, Paper, Typography } from '@material-ui/core'
import React, { useEffect } from 'react'
import { RootState, useStoreDispatch } from 'store'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import CheckCircleOutlineIcon from '@material-ui/icons/CheckCircleOutline'
import ErrorOutlineIcon from '@material-ui/icons/ErrorOutline'
import PauseCircleOutlineIcon from '@material-ui/icons/PauseCircleOutline'
import SnoozeIcon from '@material-ui/icons/Snooze'
import TimelineIcon from '@material-ui/icons/Timeline'
import { getStateofExperiments } from 'slices/experiments'
import { useSelector } from 'react-redux'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    container: {
      height: `calc(50% + ${theme.spacing(2.25)})`,
      minHeight: 100,
    },
    item: {
      display: 'flex',
      flexDirection: 'column',
      justifyContent: 'center',
      alignItems: 'center',
      height: '100%',
      textAlign: 'center',
    },
    failed: {
      color: theme.palette.error.dark,
    },
    finished: {
      color: theme.palette.success.main,
    },
  })
)

interface d {
  label: string
  value: number
  Icon: any
}

const StatusPanel = () => {
  const classes = useStyles()

  const state = useSelector((state: RootState) => state.experiments.stateOfExperiments)
  const dispatch = useStoreDispatch()

  useEffect(() => {
    dispatch(getStateofExperiments())

    const id = setInterval(() => dispatch(getStateofExperiments()), 30000)

    return () => clearInterval(id)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const data: { [k: string]: d } = {
    running: {
      label: 'Running',
      value: state.running,
      Icon: <TimelineIcon color="primary" fontSize="large" />,
    },
    paused: {
      label: 'Paused',
      value: state.paused,
      Icon: <PauseCircleOutlineIcon fontSize="large" />,
    },
    failed: {
      label: 'Failed',
      value: state.failed,
      Icon: <ErrorOutlineIcon fontSize="large" className={classes.failed} />,
    },
    waiting: {
      label: 'Waiting',
      value: state.waiting,
      Icon: <SnoozeIcon fontSize="large" />,
    },
    finished: {
      label: 'Finished',
      value: state.finished,
      Icon: <CheckCircleOutlineIcon fontSize="large" className={classes.finished} />,
    },
  }

  const StatusGrid: React.FC<{ data: d }> = ({ data }) => (
    <Grid item xs>
      <Paper variant="outlined" className={classes.item}>
        <Box display="flex" justifyContent="space-between" alignItems="center" width="50%">
          {data.Icon}
          <Box>
            <Typography color="textSecondary" gutterBottom>
              {data.label}
            </Typography>
            <Typography variant="h6">{data.value}</Typography>
          </Box>
        </Box>
      </Paper>
    </Grid>
  )

  return (
    <Box height="100%">
      <Grid container spacing={3} className={classes.container}>
        <StatusGrid data={data.running} />
        <StatusGrid data={data.paused} />
        <StatusGrid data={data.failed} />
      </Grid>
      <Grid container spacing={3} className={classes.container}>
        <StatusGrid data={data.waiting} />
        <StatusGrid data={data.finished} />
      </Grid>
    </Box>
  )
}

export default StatusPanel
