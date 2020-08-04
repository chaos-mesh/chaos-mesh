import { Box, Grid, Paper, Typography, useMediaQuery, useTheme } from '@material-ui/core'
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
      [theme.breakpoints.up('sm')]: {
        height: `calc(50% + ${theme.spacing(2.25)})`,
      },
    },
    item: {
      display: 'flex',
      alignItems: 'center',
      height: '100%',
      textAlign: 'center',
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
  const theme = useTheme()
  const isTabletScreen = useMediaQuery(theme.breakpoints.down('sm'))
  const classes = useStyles()

  const state = useSelector((state: RootState) => state.experiments.stateOfExperiments)
  const dispatch = useStoreDispatch()

  useEffect(() => {
    dispatch(getStateofExperiments())

    const id = setInterval(() => dispatch(getStateofExperiments()), 30000)

    return () => clearInterval(id)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const fontSize = isTabletScreen ? undefined : 'large'
  const data: { [k: string]: d } = {
    running: {
      label: 'Running',
      value: state.Running,
      Icon: <TimelineIcon color="primary" fontSize={fontSize} />,
    },
    paused: {
      label: 'Paused',
      value: state.Paused,
      Icon: <PauseCircleOutlineIcon color="primary" fontSize={fontSize} />,
    },
    failed: {
      label: 'Failed',
      value: state.Failed,
      Icon: <ErrorOutlineIcon color="error" fontSize={fontSize} />,
    },
    waiting: {
      label: 'Waiting',
      value: state.Waiting,
      Icon: <SnoozeIcon color="primary" fontSize={fontSize} />,
    },
    finished: {
      label: 'Finished',
      value: state.Finished,
      Icon: <CheckCircleOutlineIcon fontSize={fontSize} className={classes.finished} />,
    },
  }

  const StatusGrid: React.FC<{ data: d }> = ({ data }) => (
    <Grid item xs={6} sm>
      <Paper variant="outlined" className={classes.item}>
        <Grid container>
          <Grid item xs>
            <Box display="flex" justifyContent="center" alignItems="center" height="100%">
              {data.Icon}
            </Box>
          </Grid>
          <Grid item xs>
            <Box>
              <Typography variant="button" color="textSecondary" gutterBottom>
                {data.label}
              </Typography>
              <Typography variant="h6">{data.value}</Typography>
            </Box>
          </Grid>
        </Grid>
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
