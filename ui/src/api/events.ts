import { Event, EventsParams } from './events.type'

import http from './http'

export const events = (params?: EventsParams) =>
  http.get<Event[]>('/events', {
    params,
  })

export const get = (id: string) => http.get<Event>(`/events/${id}`)

export const cascadeFetchEventsForWorkflow = (id: string, params?: EventsParams) =>
  http.get<Event[]>(`/events/workflow/${id}`, {
    params,
  })
