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
import { type EventPageParams, useEventPage } from '@/api/events'
import Loading from '@/mui-extends/Loading'
import { useAuthStore } from '@/zustand/auth'
import { Alert, Button, Grow, Typography } from '@mui/material'
import { useState } from 'react'

import EventsTable from '@/components/EventsTable'
import NotFound from '@/components/NotFound'
import i18n from '@/components/T'

function EventsForNamespace() {
  const [params, setParams] = useState<EventPageParams>({
    page: 1,
    limit: 12,
    sort: 'created_at',
    order: 'desc',
  })
  const { data, isPending, isFetching, isError, refetch } = useEventPage(params)

  if (isPending) {
    return <Loading />
  }

  if (isError) {
    return (
      <Alert
        severity="error"
        action={
          <Button color="inherit" disabled={isFetching} onClick={() => void refetch()}>
            {i18n('events.retry')}
          </Button>
        }
      >
        {i18n('events.loadFailed')}
      </Alert>
    )
  }

  if (data.total === 0 && params.page === 1) {
    return (
      <NotFound illustrated textAlign="center">
        <Typography>{i18n('events.notFound')}</Typography>
      </NotFound>
    )
  }

  return (
    <Grow in style={{ transformOrigin: '0 0 0' }}>
      <div>
        <EventsTable
          data={data}
          order={params.order}
          orderBy={params.sort}
          pending={isFetching}
          onPageChange={(page) => setParams((current) => ({ ...current, page }))}
          onRowsPerPageChange={(limit) => setParams((current) => ({ ...current, limit, page: 1 }))}
          onSort={(sort) =>
            setParams((current) => ({
              ...current,
              sort,
              order: current.sort === sort && current.order === 'asc' ? 'desc' : 'asc',
              page: 1,
            }))
          }
        />
      </div>
    </Grow>
  )
}

export default function Events() {
  const namespace = useAuthStore((state) => state.namespace)

  return <EventsForNamespace key={namespace} />
}
