import { Box, Grid, Grow, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'

import AccountTreeOutlinedIcon from '@material-ui/icons/AccountTreeOutlined'
import { Event } from 'api/events.type'
import { Experiment } from 'api/experiments.type'
import ExperimentIcon from 'components-mui/Icons/Experiment'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import Predefined from './Predefined'
import Recent from './Recent'
import { Schedule } from 'api/schedules.type'
import ScheduleIcon from '@material-ui/icons/Schedule'
import T from 'components/T'
import TotalExperiments from './TotalExperiments'
import TotalState from './TotalState'
import Welcome from './Welcome'
import { Workflow } from 'api/workflows.type'
import api from 'api'
import { makeStyles } from '@material-ui/styles'

const useStyles = makeStyles(() => ({
  totalState: {
    position: 'relative',
    height: 300,
  },
}))

const NumPanel: React.FC<{ title: JSX.Element; num: number; background: JSX.Element }> = ({
  title,
  num,
  background,
}) => (
  <Paper boxProps={{ overflow: 'hidden' }}>
    <PaperTop title={title} />
    <Box mt={6}>
      <Typography component="div" variant="h4">
        {num}
      </Typography>
    </Box>
    <Box position="absolute" bottom={-18} right={12}>
      {background}
    </Box>
  </Paper>
)

export default function Dashboard() {
  const classes = useStyles()

  const [data, setData] = useState<{
    workflows: Workflow[]
    schedules: Schedule[]
    experiments: Experiment[]
    events: Event[]
  }>({
    workflows: [],
    schedules: [],
    experiments: [],
    events: [],
  })

  useEffect(() => {
    const fetchExperiments = api.experiments.experiments()
    const fetchSchedules = api.schedules.schedules()
    const fetchWorkflows = api.workflows.workflows()
    const fetchEvents = api.events.events()
    const fetchAll = () => {
      Promise.all([fetchSchedules, fetchWorkflows, fetchExperiments, fetchEvents])
        .then((data) =>
          setData({
            schedules: data[0].data,
            workflows: data[1].data,
            experiments: data[2].data,
            events: data[3].data,
          })
        )
        .catch(console.error)
    }

    fetchAll()

    // const id = setInterval(fetchEvents, 15000)

    // return () => clearInterval(id)
  }, [])

  return (
    <Grow in={true} style={{ transformOrigin: '0 0 0' }}>
      <Grid container spacing={6}>
        <Grid container spacing={6} alignContent="flex-start" item xs={12} lg={8}>
          <Grid item xs={4}>
            <NumPanel
              title={T('experiments.title')}
              num={data.experiments.length}
              background={<ExperimentIcon color="primary" style={{ fontSize: '3em' }} />}
            />
          </Grid>
          <Grid item xs={4}>
            <NumPanel
              title={T('schedules.title')}
              num={data.schedules.length}
              background={<ScheduleIcon color="primary" style={{ fontSize: '3em' }} />}
            />
          </Grid>
          <Grid item xs={4}>
            <NumPanel
              title={T('workflows.title')}
              num={data.workflows.length}
              background={<AccountTreeOutlinedIcon color="primary" style={{ fontSize: '3em' }} />}
            />
          </Grid>
          <Grid item xs={12}>
            <Welcome />
          </Grid>
          <Grid item xs={12}>
            <Paper>
              <PaperTop title={T('dashboard.predefined')} subtitle={T('dashboard.predefinedDesc')} />
              <Predefined />
            </Paper>
          </Grid>
        </Grid>

        <Grid container spacing={6} item xs={12} lg={4}>
          <Grid item xs={12}>
            <Paper>
              <PaperTop title={T('dashboard.totalState')} />
              <TotalState className={classes.totalState} />
            </Paper>
          </Grid>
          <Grid item xs={12}>
            <Paper>
              <PaperTop title={T('dashboard.recent')} />
              <Recent events={data.events.slice(-6)} />
            </Paper>
          </Grid>
        </Grid>
      </Grid>
    </Grow>
  )
}
