import { Box, Grow, Paper, Typography } from '@material-ui/core'
import React, { useEffect, useRef, useState } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import BlurLinearIcon from '@material-ui/icons/BlurLinear'
import ContentContainer from 'components/ContentContainer'
import { Event } from 'api/events.type'
import EventsTable from 'components/EventsTable'
import Loading from 'components/Loading'
import PaperTop from 'components/PaperTop'
import api from 'api'
import clsx from 'clsx'
import genEventsChart from 'lib/d3/eventsChart'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    height100: {
      [theme.breakpoints.up('md')]: {
        height: '100%',
      },
    },
    timelinePaper: {
      marginBottom: theme.spacing(3),
    },
    eventsChart: {
      height: 300,
    },
    paper: {
      padding: theme.spacing(3),
      paddingTop: 0,
    },
  })
)

export default function Events() {
  const classes = useStyles()

  const chartRef = useRef<HTMLDivElement>(null)
  const [loading, setLoading] = useState(false)
  const [events, setEvents] = useState<Event[] | null>(null)

  const fetchEvents = () => {
    setLoading(true)

    api.events
      .events()
      .then(({ data }) => setEvents(data))
      .catch(console.log)
      .finally(() => {
        setLoading(false)
      })
  }

  useEffect(fetchEvents, [])

  useEffect(() => {
    if (events && events.length) {
      const chart = chartRef.current!

      genEventsChart({
        root: chart,
        events,
      })
    }
  }, [events])

  return (
    <ContentContainer>
      {events && events.length > 0 && (
        <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
          <Box display="flex" flexDirection="column" height="100%">
            <Paper className={clsx(classes.paper, classes.timelinePaper)}>
              <PaperTop title="Timeline" />
              <div ref={chartRef} className={classes.eventsChart} />
            </Paper>
            <Paper className={clsx(classes.height100, classes.paper)}>
              <EventsTable events={events} detailed />
            </Paper>
          </Box>
        </Grow>
      )}

      {!loading && events && events.length === 0 && (
        <Box display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="100%">
          <Box mb={3}>
            <BlurLinearIcon fontSize="large" />
          </Box>
          <Typography variant="h6" align="center">
            No events found. Try to create a new experiment.
          </Typography>
        </Box>
      )}

      {loading && <Loading />}
    </ContentContainer>
  )
}
