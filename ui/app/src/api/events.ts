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
import type { CoreEvent } from '@/openapi/index.schemas'
import { useAuthStore } from '@/zustand/auth'
import { useQuery } from '@tanstack/react-query'

import { customInstance } from './http'

export type EventSort = 'object_id' | 'namespace' | 'name' | 'kind' | 'created_at' | 'message'

export interface EventPageParams {
  page: number
  limit: number
  sort: EventSort
  order: 'asc' | 'desc'
}

export interface EventPage {
  events: CoreEvent[]
  page: number
  limit: number
  next_page: number | null
  total: number
  total_pages: number
}

export function useEventPage(params: EventPageParams) {
  const namespace = useAuthStore((state) => state.namespace)

  return useQuery({
    queryKey: ['/events', 'page', namespace, params],
    queryFn: ({ signal }) =>
      customInstance<EventPage>({
        url: '/events',
        method: 'GET',
        params: { ...params, namespace: namespace === 'All' ? undefined : namespace, include_total: true },
        signal,
      }),
    retry: 1,
    retryDelay: 3000,
  })
}
