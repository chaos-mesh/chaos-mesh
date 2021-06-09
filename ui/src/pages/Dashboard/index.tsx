import { Box, Grid, Grow, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'

import AccountTreeOutlinedIcon from '@material-ui/icons/AccountTreeOutlined'
import { Event } from 'api/events.type'
import EventsChart from 'components/EventsChart'
import EventsTimeline from 'components/EventsTimeline'
import { Experiment } from 'api/experiments.type'
import ExperimentIcon from 'components-mui/Icons/Experiment'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import Predefined from './Predefined'
import { Schedule } from 'api/schedules.type'
import ScheduleIcon from '@material-ui/icons/Schedule'
import T from 'components/T'
import TotalState from './TotalState'
import Welcome from './Welcome'
import { Workflow } from 'api/workflows.type'
import api from 'api'
import { makeStyles } from '@material-ui/styles'

const useStyles = makeStyles({
  container: {
    position: 'relative',
    height: 300,
  },
})

const NumPanel: React.FC<{ title: JSX.Element; num: number; background: JSX.Element }> = ({
  title,
  num,
  background,
}) => (
  <Paper sx={{ overflow: 'hidden' }}>
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
    const fetchEvents = api.events.events({ limit: 6 })
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

    const id = setInterval(fetchAll, 12000)

    return () => clearInterval(id)
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
              <PaperTop
                title={T('dashboard.predefined')}
                subtitle={T('dashboard.predefinedDesc')}
                boxProps={{ mb: 3 }}
              />
              <Predefined />
            </Paper>
          </Grid>
          <Grid item xs={12}>
            <Paper>
              <PaperTop title={T('common.timeline')} boxProps={{ mb: 3 }} />
              <EventsChart events={data.events} className={classes.container} />
            </Paper>
          </Grid>
        </Grid>

        <Grid container spacing={6} item xs={12} lg={4}>
          <Grid item xs={12}>
            <Paper>
              <PaperTop title={T('dashboard.totalState')} />
              <TotalState className={classes.container} />
            </Paper>
          </Grid>
          <Grid item xs={12}>
            <Paper>
              <PaperTop title={T('dashboard.recent')} boxProps={{ mb: 3 }} />
              <EventsTimeline events={data.events} />
            </Paper>
          </Grid>
        </Grid>
      </Grid>
    </Grow>
  )
}
