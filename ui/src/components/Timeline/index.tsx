import React, { useEffect, useRef } from 'react'

import { Event } from 'api/events.type'
import T from 'components/T'
import { Typography } from '@material-ui/core'
import genEventsChart from 'lib/d3/eventsChart'
import { makeStyles } from '@material-ui/core/styles'
import { useIntl } from 'react-intl'
import { useStoreSelector } from 'store'

const useStyles = makeStyles({
  notFound: {
    position: 'absolute',
    top: '50%',
    left: '50%',
    transform: 'translate3d(-50%, -50%, 0)',
  },
})

interface TimelineProps {
  events: Event[]
  className: string
}

const Timeline: React.FC<TimelineProps> = ({ events, ...rest }) => {
  const classes = useStyles()

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
      {events?.length === 0 && <Typography className={classes.notFound}>{T('events.noEventsFound')}</Typography>}
    </div>
  )
}

export default Timeline
