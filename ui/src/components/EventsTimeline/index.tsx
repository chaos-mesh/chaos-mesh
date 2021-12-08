/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import { Box, Typography } from '@material-ui/core'
import DateTime, { format } from 'lib/luxon'

import { Event } from 'api/events.type'
import NotFound from 'components-mui/NotFound'
import T from 'components/T'
import Timeline from '@material-ui/lab/Timeline'
import TimelineConnector from '@material-ui/lab/TimelineConnector'
import TimelineContent from '@material-ui/lab/TimelineContent'
import TimelineDot from '@material-ui/lab/TimelineDot'
import TimelineItem from '@material-ui/lab/TimelineItem'
import TimelineOppositeContent from '@material-ui/lab/TimelineOppositeContent'
import TimelineSeparator from '@material-ui/lab/TimelineSeparator'
import { iconByKind } from 'lib/byKind'
import { useStoreSelector } from 'store'

interface EventsTimelineProps {
  events: Event[]
}

const EventsTimeline: React.FC<EventsTimelineProps> = ({ events }) => {
  const { lang } = useStoreSelector((state) => state.settings)

  return events.length > 0 ? (
    <Timeline sx={{ m: 0, p: 0 }}>
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
  ) : (
    <NotFound>{T('events.notFound')}</NotFound>
  )
}

export default EventsTimeline
