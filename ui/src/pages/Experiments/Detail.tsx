import { Box, Divider, Grid, Grow, Paper, Typography } from '@material-ui/core'
import React, { useEffect, useRef, useState } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'
import { useLocation, useParams } from 'react-router-dom'

import ContentContainer from 'components/ContentContainer'
import ErrorOutlineIcon from '@material-ui/icons/ErrorOutline'
import { Event } from 'api/events.type'
import { Experiment } from 'components/NewExperiment/types'
import Loading from 'components/Loading'
import ReactJson from 'react-json-view'
import api from 'api'
import genEventsChart from 'lib/eventsChart'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    height100: {
      [theme.breakpoints.up('md')]: {
        height: '100%',
      },
    },
    eventsChart: {
      height: 300,
      marginBottom: theme.spacing(3),
    },
    experimentPaper: {
      flex: 1,
      padding: theme.spacing(3),
    },
  })
)

export default function ExperimentDetail() {
  const classes = useStyles()

  const { search } = useLocation()
  const searchParams = new URLSearchParams(search)
  const namespace = searchParams.get('namespace')
  const kind = searchParams.get('kind')
  const { name } = useParams()

  const chartRef = useRef<HTMLDivElement>(null)
  const [loading, setLoading] = useState(false)
  const [detail, setDetail] = useState<Experiment | null>(null)
  const [events, setEvents] = useState<Event[] | null>(null)
  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null)

  const fetchExperimentDetail = () => {
    setLoading(true)

    api.experiments
      .detail(namespace!, name!, kind!)
      .then(({ data }) => setDetail(data))
      .catch(console.log)
  }

  const fetchEventsByName = (name: string) => {
    api.events
      .events()
      .then(({ data }) => {
        setEvents(data.filter((d) => d.Experiment === name))
      })
      .catch(console.log)
      .finally(() => {
        setLoading(false)
      })
  }

  useEffect(() => {
    if (namespace && kind && name) {
      fetchExperimentDetail()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [namespace, kind, name])

  useEffect(() => {
    if (detail) {
      fetchEventsByName(detail.name)
    }
  }, [detail])

  useEffect(() => {
    if (events) {
      setSelectedEvent(events[events.length - 1])

      const chart = chartRef.current!

      genEventsChart({
        root: chart,
        width: chart.offsetWidth,
        height: chart.offsetHeight,
        events,
        selectEvent: setSelectedEvent,
      })
    }
  }, [events])

  return (
    <ContentContainer>
      <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
        <Grid className={classes.height100} container spacing={3}>
          <Grid item xs={12} sm={12} md={8}>
            <Box display="flex" flexDirection="column" height="100%">
              <Paper ref={chartRef} className={classes.eventsChart}></Paper>
              <Paper className={`${classes.height100} ${classes.experimentPaper}`}>
                <Box pb={3}>
                  <Typography variant="h6">Event</Typography>
                </Box>
                <Box pb={3}>
                  <Divider />
                </Box>
                {selectedEvent && <ReactJson src={selectedEvent} collapsed={1} />}
              </Paper>
            </Box>
          </Grid>
          <Grid item xs={12} sm={12} md={4}>
            <Paper className={`${classes.height100} ${classes.experimentPaper}`}>
              <Box pb={3}>
                <Typography variant="h6">Configuration</Typography>
              </Box>
              <Box pb={3}>
                <Divider />
              </Box>
              {detail && <ReactJson src={detail} collapsed={1} />}
            </Paper>
          </Grid>
        </Grid>
      </Grow>

      {(!namespace || !kind || !name) && (
        <Box display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="100%">
          <ErrorOutlineIcon fontSize="large" />
          <Typography variant="h6" align="center">
            Please check the URL params and queries to provide the correct params.
          </Typography>
        </Box>
      )}

      {loading && <Loading />}
    </ContentContainer>
  )
}
