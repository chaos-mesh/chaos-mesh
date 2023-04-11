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
import Timeline from '@mui/lab/Timeline'
import TimelineConnector from '@mui/lab/TimelineConnector'
import TimelineContent from '@mui/lab/TimelineContent'
import TimelineDot from '@mui/lab/TimelineDot'
import TimelineItem from '@mui/lab/TimelineItem'
import TimelineOppositeContent from '@mui/lab/TimelineOppositeContent'
import TimelineSeparator from '@mui/lab/TimelineSeparator'
import { Box, Typography } from '@mui/material'
import { CoreEvent } from 'openapi/index.schemas'

import { useStoreSelector } from 'store'

import NotFound from 'components/NotFound'
import i18n from 'components/T'

import { iconByKind } from 'lib/byKind'
import DateTime, { format } from 'lib/luxon'

interface EventsTimelineProps {
  events: CoreEvent[]
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
            <TimelineDot color="primary">{iconByKind(e.kind as any, 'small')}</TimelineDot>
          </TimelineSeparator>
          <TimelineContent>
            <Box display="flex" justifyContent="space-between" mt={6}>
              <Box flex={1} ml={3}>
                <Typography gutterBottom>{e.name}</Typography>
                <Typography variant="body2" color="textSecondary">
                  {e.message}
                </Typography>
              </Box>
              <Typography variant="overline" title={format(e.created_at!)}>
                {DateTime.fromISO(e.created_at!, {
                  locale: lang,
                }).toRelative()}
              </Typography>
            </Box>
          </TimelineContent>
        </TimelineItem>
      ))}
    </Timeline>
  ) : (
    <NotFound>{i18n('events.notFound')}</NotFound>
  )
}

export default EventsTimeline
