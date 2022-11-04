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
import { Box, BoxProps } from '@mui/material'
import { CoreEvent } from 'openapi/index.schemas'
import { useEffect, useRef } from 'react'

import { useStoreSelector } from 'store'

import NotFound from 'components/NotFound'
import i18n from 'components/T'

import genEventsChart from 'lib/d3/eventsChart'

interface EventsChartProps extends BoxProps {
  events: CoreEvent[]
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
      {events.length === 0 && <NotFound>{i18n('events.notFound')}</NotFound>}
    </Box>
  )
}

export default EventsChart
