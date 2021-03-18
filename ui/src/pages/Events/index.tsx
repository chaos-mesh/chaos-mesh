import { Box, Grow, Typography } from '@material-ui/core'
import EventsTable, { EventsTableHandles } from 'components/EventsTable'
import React, { useEffect, useRef, useState } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import { Event } from 'api/events.type'
import Loading from 'components-mui/Loading'
import NotFound from 'components-mui/NotFound'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import { RootState } from 'store'
import T from 'components/T'
import TimelineOutlinedIcon from '@material-ui/icons/TimelineOutlined'
import api from 'api'
import genEventsChart from 'lib/d3/eventsChart'
import { useIntl } from 'react-intl'
import { useSelector } from 'react-redux'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    timelinePaper: {
      marginBottom: theme.spacing(6),
    },
    eventsChart: {
      height: 450,
      margin: theme.spacing(3),
    },
  })
)

export default function Events() {
  const classes = useStyles()

  const intl = useIntl()

  const { theme } = useSelector((state: RootState) => state.settings)

  const chartRef = useRef<HTMLDivElement>(null)
  const eventsTableRef = useRef<EventsTableHandles>(null)

  const [loading, setLoading] = useState(false)
  const [events, setEvents] = useState<Event[] | null>(null)

  const fetchEvents = () => {
    setLoading(true)

    api.events
      .events()
      .then(({ data }) => setEvents(data))
      .catch(console.error)
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
        intl,
        theme,
        options: {
          enableLegends: true,
          onSelectEvent: eventsTableRef.current!.onSelectEvent,
        },
      })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [events])

  return (
    <>
      {events && events.length > 0 && (
        <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
          <Box display="flex" flexDirection="column">
            <Paper className={classes.timelinePaper}>
              <PaperTop title={T('common.timeline')} />
              <div ref={chartRef} className={classes.eventsChart} />
            </Paper>
            <EventsTable ref={eventsTableRef} events={events} detailed />
          </Box>
        </Grow>
      )}

      {!loading && events && events.length === 0 && (
        <NotFound textAlign="center">
          <Box mb={3}>
            <TimelineOutlinedIcon fontSize="large" />
          </Box>
          <Typography variant="h6">{T('events.noEventsFound')}</Typography>
        </NotFound>
      )}

      {loading && <Loading />}
    </>
  )
}
