import { Grow, Typography } from '@material-ui/core'
import { useEffect, useState } from 'react'

import { Event } from 'api/events.type'
import EventsTable from 'components/EventsTable'
import Loading from 'components-mui/Loading'
import NotFound from 'components-mui/NotFound'
import T from 'components/T'
import api from 'api'

export default function Events() {
  const [loading, setLoading] = useState(true)
  const [events, setEvents] = useState<Event[]>([])

  useEffect(() => {
    const fetchEvents = () => {
      api.events
        .events()
        .then(({ data }) => setEvents(data))
        .catch(console.error)
        .finally(() => setLoading(false))
    }

    fetchEvents()
  }, [])

  return (
    <>
      {events && events.length > 0 && (
        <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
          <div>
            <EventsTable events={events} />
          </div>
        </Grow>
      )}

      {!loading && events.length === 0 && (
        <NotFound illustrated textAlign="center">
          <Typography>{T('events.notFound')}</Typography>
        </NotFound>
      )}

      {loading && <Loading />}
    </>
  )
}
