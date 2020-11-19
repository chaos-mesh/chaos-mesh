import { Event, EventsParams } from './events.type'

import http from './http'

export const events = (params?: EventsParams) =>
  http.get<Event[]>('/events', {
    params,
  })

// Without pods field
export const dryEvents = (params?: EventsParams) =>
  http.get<Event[]>('/events/dry', {
    params,
  })
