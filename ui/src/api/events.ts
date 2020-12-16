import { Event, EventsParams } from './events.type'

import http from './http'

export const events = (namespace = '') =>
  http.get<Event[]>('/events', {
    params: {
      namespace,
    },
  })

// Without pods field
export const dryEvents = (namespace = '') =>
  http.get<Event[]>('/events/dry', {
    params: {
      namespace,
    },
  })
