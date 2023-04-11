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
import { Grow, Typography } from '@mui/material'
import { useGetEvents } from 'openapi'

import Loading from '@ui/mui-extends/esm/Loading'

import EventsTable from 'components/EventsTable'
import NotFound from 'components/NotFound'
import i18n from 'components/T'

export default function Events() {
  const { data: events, isLoading: loading } = useGetEvents()

  return (
    <>
      {events && events.length > 0 && (
        <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
          <div>
            <EventsTable events={events} />
          </div>
        </Grow>
      )}

      {!loading && events?.length === 0 && (
        <NotFound illustrated textAlign="center">
          <Typography>{i18n('events.notFound')}</Typography>
        </NotFound>
      )}

      {loading && <Loading />}
    </>
  )
}
