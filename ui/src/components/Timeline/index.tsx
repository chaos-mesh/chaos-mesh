import React, { useEffect, useRef, useState } from 'react'

import { Event } from 'api/events.type'
import api from 'api'
import genEventsChart from 'lib/d3/eventsChart'
import { useIntl } from 'react-intl'
import { useStoreSelector } from 'store'

interface TimelineProps {
  className: string
}

const Timeline: React.FC<TimelineProps> = (props) => {
  const intl = useIntl()

  const { theme } = useStoreSelector((state) => state.settings)

  const chartRef = useRef<HTMLDivElement>(null)

  const [events, setEvents] = useState<Event[] | null>(null)

  const fetchEvents = () => {
    api.events
      .events()
      .then(({ data }) => setEvents(data))
      .catch(console.error)
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
      })
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [events])

  return <div ref={chartRef} {...props} />
}

export default Timeline
