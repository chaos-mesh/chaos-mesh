import { Box, Grid, Paper, Typography, useMediaQuery, useTheme } from '@material-ui/core'
import React, { useEffect } from 'react'
import { RootState, useStoreDispatch } from 'store'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import CheckCircleOutlineIcon from '@material-ui/icons/CheckCircleOutline'
import ErrorOutlineIcon from '@material-ui/icons/ErrorOutline'
import PauseCircleOutlineIcon from '@material-ui/icons/PauseCircleOutline'
import SnoozeIcon from '@material-ui/icons/Snooze'
import T from 'components/T'
import TimelineIcon from '@material-ui/icons/Timeline'
import { getStateofExperiments } from 'slices/experiments'
import { useSelector } from 'react-redux'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    container: {
      [theme.breakpoints.up('sm')]: {
        height: `calc(100% + ${theme.spacing(3)})`,
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
      label: 'running',
      value: state.Running,
      Icon: <TimelineIcon color="primary" fontSize={fontSize} />,
    },
    paused: {
      label: 'paused',
      value: state.Paused,
      Icon: <PauseCircleOutlineIcon color="primary" fontSize={fontSize} />,
    },
    failed: {
      label: 'failed',
      value: state.Failed,
      Icon: <ErrorOutlineIcon color="error" fontSize={fontSize} />,
    },
    waiting: {
      label: 'waiting',
      value: state.Waiting,
      Icon: <SnoozeIcon color="primary" fontSize={fontSize} />,
    },
    finished: {
      label: 'finished',
      value: state.Finished,
      Icon: <CheckCircleOutlineIcon fontSize={fontSize} className={classes.finished} />,
    },
  }

  const StatusGrid: React.FC<{ data: d; sm: any }> = ({ data, sm }) => (
    <Grid item xs={6} sm={sm}>
      <Paper variant="outlined" className={classes.item}>
        <Grid container>
          <Grid item xs>
            <Box display="flex" justifyContent="center" alignItems="center" height="100%">
              {data.Icon}
            </Box>
          </Grid>
          <Grid item xs>
            <Box>
              <Typography variant="overline">{T(`experiments.status.${data.label}`)}</Typography>
              <Typography variant="h5">{data.value}</Typography>
            </Box>
          </Grid>
        </Grid>
      </Paper>
    </Grid>
  )

  return (
    <Box height="100%">
      <Grid container spacing={3} className={classes.container}>
        <StatusGrid data={data.running} sm={4} />
        <StatusGrid data={data.paused} sm={4} />
        <StatusGrid data={data.failed} sm={4} />
        <StatusGrid data={data.waiting} sm={6} />
        <StatusGrid data={data.finished} sm={6} />
      </Grid>
    </Box>
  )
}

export default StatusPanel
