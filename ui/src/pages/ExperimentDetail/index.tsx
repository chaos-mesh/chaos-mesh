import { Box, Grid, Grow, IconButton, Paper, Typography } from '@material-ui/core'
import React, { useEffect, useRef, useState } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'
import { useLocation, useParams } from 'react-router-dom'

import CloseIcon from '@material-ui/icons/Close'
import ContentContainer from 'components/ContentContainer'
import ErrorOutlineIcon from '@material-ui/icons/ErrorOutline'
import { Event } from 'api/events.type'
import EventDetail from 'components/EventDetail'
import EventsTable from 'components/EventsTable'
import { Experiment } from 'components/NewExperiment/types'
import Loading from 'components/Loading'
import PageTitle from 'components/PageTitle'
import ReactJson from 'react-json-view'
import api from 'api'
import clsx from 'clsx'
import genEventsChart from 'lib/d3/eventsChart'
import { usePrevious } from 'lib/hooks'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    height100: {
      [theme.breakpoints.up('md')]: {
        height: '100%',
      },
    },
    timelinePaper: {
      marginBottom: theme.spacing(3),
      padding: theme.spacing(3),
    },
    eventsChart: {
      height: 300,
    },
    paper: {
      padding: theme.spacing(3),
    },
    detailPaper: {
      position: 'absolute',
      top: 0,
      left: 0,
      width: '100%',
      height: '100%',
      overflow: 'scroll',
    },
  })
)

export default function ExperimentDetail() {
  const classes = useStyles()

  const { search } = useLocation()
  const searchParams = new URLSearchParams(search)
  const namespace = searchParams.get('namespace')
  const kind = searchParams.get('kind')
  const eventID = searchParams.get('event')
  const prevEventID = usePrevious(eventID)
  const { name } = useParams()

  const chartRef = useRef<HTMLDivElement>(null)
  const [loading, setLoading] = useState(false)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detail, setDetail] = useState<Experiment | null>(null)
  const [events, setEvents] = useState<Event[] | null>(null)
  const prevEvents = usePrevious(events)
  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null)
  const [eventDetailOpen, setEventDetailOpen] = useState(false)

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
      .then(({ data }) => setEvents(data.filter((d) => d.Experiment === name)))
      .catch(console.log)
      .finally(() => {
        setLoading(false)
      })
  }

  const onSelectEvent = (e: Event) => {
    setDetailLoading(true)
    setSelectedEvent(e)
    setEventDetailOpen(true)
    setTimeout(() => setDetailLoading(false), 1000)
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
    if (prevEvents !== events && events) {
      const chart = chartRef.current!

      genEventsChart({
        root: chart,
        events,
        selectEvent: onSelectEvent,
      })

      if (eventID !== null) {
        onSelectEvent(events.filter((e) => e.ID === parseInt(eventID))[0])
      }
    }

    if (prevEventID !== eventID && eventID !== null && events) {
      onSelectEvent(events.filter((e) => e.ID === parseInt(eventID))[0])
    }
  }, [prevEvents, events, prevEventID, eventID])

  return (
    <ContentContainer>
      <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
        <Grid className={classes.height100} container spacing={3}>
          <Grid item xs={12} sm={12} md={9}>
            <Box display="flex" flexDirection="column" height="100%">
              <Paper className={classes.timelinePaper}>
                <PageTitle>Timeline</PageTitle>
                <div ref={chartRef} className={classes.eventsChart} />
              </Paper>
              <Box className={classes.height100} position="relative">
                <Paper className={clsx(classes.height100, classes.paper)}>
                  <PageTitle>Events</PageTitle>
                  <EventsTable events={events} detailed />
                </Paper>
                {eventDetailOpen && (
                  <Paper className={clsx(classes.paper, classes.detailPaper)}>
                    <Box display="flex" justifyContent="space-between">
                      <PageTitle>Event</PageTitle>
                      <IconButton color="primary" onClick={() => setEventDetailOpen(false)}>
                        <CloseIcon />
                      </IconButton>
                    </Box>
                    {selectedEvent && !detailLoading ? (
                      <Box ml={3} mb={3}>
                        <EventDetail event={selectedEvent} />
                      </Box>
                    ) : (
                      <Loading />
                    )}
                  </Paper>
                )}
              </Box>
            </Box>
          </Grid>
          <Grid item xs={12} sm={12} md={3}>
            <Paper className={clsx(classes.height100, classes.paper)}>
              <PageTitle>Configuration</PageTitle>
              {detail && (
                <Box ml={3}>
                  <ReactJson src={detail} collapsed={1} displayObjectSize={false} />
                </Box>
              )}
            </Paper>
          </Grid>
        </Grid>
      </Grow>

      {(!namespace || !kind || !name) && (
        <Box display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="100%">
          <Box mb={3}>
            <ErrorOutlineIcon fontSize="large" />
          </Box>
          <Typography variant="h6" align="center">
            Please check the URL params and queries to provide the correct params.
          </Typography>
        </Box>
      )}

      {loading && <Loading />}
    </ContentContainer>
  )
}
