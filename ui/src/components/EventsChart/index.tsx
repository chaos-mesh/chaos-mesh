import { Box, BoxProps } from '@material-ui/core'
import { useEffect, useRef } from 'react'

import { Event } from 'api/events.type'
import NotFound from 'components-mui/NotFound'
import T from 'components/T'
import genEventsChart from 'lib/d3/eventsChart'
import { useStoreSelector } from 'store'

interface EventsChartProps extends BoxProps {
  events: Event[]
}

const EventsChart: React.FC<EventsChartProps> = ({ events, ...rest }) => {
  const { theme } = useStoreSelector((state) => state.settings)

  const chartRef = useRef<any>(null)

  useEffect(() => {
    if (events.length) {
      const chart = chartRef.current!

      if (typeof chart === 'function') {
        chart(events)

        return
      }

      chartRef.current = genEventsChart({
        root: chart,
        events,
        theme,
      })
    }
  }, [events, theme])

  return (
    <Box {...rest} ref={chartRef}>
      {events?.length === 0 && <NotFound>{T('events.notFound')}</NotFound>}
    </Box>
  )
}

export default EventsChart
