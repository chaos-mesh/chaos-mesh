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
import Paper from '@/mui-extends/Paper'
import PaperTop from '@/mui-extends/PaperTop'
import { type CoreEvent } from '@/openapi/index.schemas'
import { useSettingActions, useSettingStore } from '@/zustand/setting'
import { useSystemStore } from '@/zustand/system'
import Timeline from '@mui/lab/Timeline'
import TimelineConnector from '@mui/lab/TimelineConnector'
import TimelineContent from '@mui/lab/TimelineContent'
import TimelineDot from '@mui/lab/TimelineDot'
import TimelineItem from '@mui/lab/TimelineItem'
import TimelineOppositeContent from '@mui/lab/TimelineOppositeContent'
import TimelineSeparator from '@mui/lab/TimelineSeparator'
import { Box, FormControlLabel, Switch, Typography } from '@mui/material'

import NotFound from '@/components/NotFound'
import i18n from '@/components/T'

import { iconByKind } from '@/lib/byKind'
import { format, toRelative } from '@/lib/luxon'

interface EventsTimelineProps {
  events: CoreEvent[] | undefined
  paperProps?: React.ComponentProps<typeof Paper>
}

const EventsTimeline: ReactFCWithChildren<EventsTimelineProps> = ({ events, paperProps }) => {
  const lang = useSystemStore((state) => state.lang)
  const eventTimeFormat = useSettingStore((state) => state.eventTimeFormat)
  const { setEventTimeFormat } = useSettingActions()

  const handleEventTimeFormatChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setEventTimeFormat(event.target.checked ? 'absolute' : 'relative')
  }

  return (
    <Paper {...paperProps} sx={{ display: 'flex', flexDirection: 'column', ...paperProps?.sx }}>
      <PaperTop title={paperProps?.title || i18n('events.title')} boxProps={{ mb: 3 }}>
        <FormControlLabel
          control={
            <Switch size="small" checked={eventTimeFormat === 'absolute'} onChange={handleEventTimeFormatChange} />
          }
          label={i18n('events.absoluteTime')}
          sx={{ mr: 0 }}
        />
      </PaperTop>
      <Box flex={1} overflow="scroll">
        {events && events.length > 0 ? (
          <Timeline sx={{ m: 0, p: 0 }}>
            {events.map((e) => (
              <TimelineItem key={e.id}>
                <TimelineOppositeContent style={{ flex: 0, padding: 0 }} />
                <TimelineSeparator>
                  <TimelineConnector />
                  <TimelineDot color="primary">{iconByKind(e.kind!, 'small')}</TimelineDot>
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
                      {eventTimeFormat === 'absolute' ? format(e.created_at!, lang) : toRelative(e.created_at!, lang)}
                    </Typography>
                  </Box>
                </TimelineContent>
              </TimelineItem>
            ))}
          </Timeline>
        ) : (
          <NotFound>{i18n('events.notFound')}</NotFound>
        )}
      </Box>
    </Paper>
  )
}

export default EventsTimeline
