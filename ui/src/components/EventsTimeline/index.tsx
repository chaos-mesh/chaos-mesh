import { Box, Typography } from '@material-ui/core'
import DateTime, { format } from 'lib/luxon'

import { Event } from 'api/events.type'
import Timeline from '@material-ui/lab/Timeline'
import TimelineConnector from '@material-ui/lab/TimelineConnector'
import TimelineContent from '@material-ui/lab/TimelineContent'
import TimelineDot from '@material-ui/lab/TimelineDot'
import TimelineItem from '@material-ui/lab/TimelineItem'
import TimelineOppositeContent from '@material-ui/lab/TimelineOppositeContent'
import TimelineSeparator from '@material-ui/lab/TimelineSeparator'
import { iconByKind } from 'lib/byKind'
import { makeStyles } from '@material-ui/styles'
import { useStoreSelector } from 'store'

const useStyles = makeStyles({
  root: {
    margin: 0,
    padding: 0,
  },
})

interface EventsTimelineProps {
  events: Event[]
}

const EventsTimeline: React.FC<EventsTimelineProps> = ({ events }) => {
  const classes = useStyles()

  const { lang } = useStoreSelector((state) => state.settings)

  return (
    <Timeline classes={classes}>
      {events.map((e) => (
        <TimelineItem key={e.id}>
          <TimelineOppositeContent style={{ flex: 0.001, padding: 0 }} />
          <TimelineSeparator>
            <TimelineConnector sx={{ py: 3 }} />
            <TimelineDot color="primary">{iconByKind(e.kind, 'small')}</TimelineDot>
          </TimelineSeparator>
          <TimelineContent>
            <Box display="flex" justifyContent="space-between" mt={6}>
              <Box flex={1} ml={3}>
                <Typography gutterBottom>{e.name}</Typography>
                <Typography variant="body2" color="textSecondary">
                  {e.message}
                </Typography>
              </Box>
              <Typography variant="overline" title={format(e.created_at)}>
                {DateTime.fromISO(e.created_at, {
                  locale: lang,
                }).toRelative()}
              </Typography>
            </Box>
          </TimelineContent>
        </TimelineItem>
      ))}
    </Timeline>
  )
}

export default EventsTimeline
