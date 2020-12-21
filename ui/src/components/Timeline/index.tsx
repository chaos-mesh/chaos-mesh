import React, { useEffect, useRef } from 'react'

import { Event } from 'api/events.type'
import NotFound from 'components-mui/NotFound'
import T from 'components/T'
import genEventsChart from 'lib/d3/eventsChart'
import { useIntl } from 'react-intl'
import { useStoreSelector } from 'store'

interface TimelineProps {
  events: Event[]
  className: string
}

const Timeline: React.FC<TimelineProps> = ({ events, ...rest }) => {
  const intl = useIntl()

  const { theme } = useStoreSelector((state) => state.settings)

  const chartRef = useRef<HTMLDivElement>(null)

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

  return (
    <div ref={chartRef} {...rest}>
      {events?.length === 0 && <NotFound>{T('events.noEventsFound')}</NotFound>}
    </div>
  )
}

export default Timeline
