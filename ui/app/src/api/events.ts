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
import { useGetEvents, useGetEventsPage } from '@/openapi'
import type { EventPageResponse, GetEventsPageParams, GetEventsPageSort } from '@/openapi/index.schemas'
import { useAuthStore } from '@/zustand/auth'

interface EventListParams {
  object_id?: string
  limit?: number
}

function useEventNamespace() {
  const namespace = useAuthStore((state) => state.namespace)
  return namespace === 'All' ? undefined : [namespace]
}

export function useEventList(params?: EventListParams, enabled = true) {
  const namespace = useEventNamespace()
  const objectID = params?.object_id

  return useGetEvents(
    { limit: params?.limit, object_id: objectID ? [objectID] : undefined, namespace },
    { query: { enabled } },
  )
}

export type EventSort = Exclude<GetEventsPageSort, 'id'>

export type EventPageParams = Required<Pick<GetEventsPageParams, 'page' | 'limit' | 'order'>> & { sort: EventSort }

export type EventPage = EventPageResponse & Required<Pick<EventPageResponse, 'total' | 'total_pages'>>

export function useEventPage(params: EventPageParams) {
  const namespace = useEventNamespace()

  return useGetEventsPage(
    { ...params, namespace, include_total: true },
    {
      query: {
        select: (page): EventPage => {
          const { total, total_pages } = page
          if (total === undefined || total_pages === undefined) {
            throw new Error('Event page response is missing totals')
          }
          return { ...page, total, total_pages }
        },
      },
    },
  )
}
