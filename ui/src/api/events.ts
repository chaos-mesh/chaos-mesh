import { Event, EventsParams } from './events.type'

import http from './http'

export const events = (params?: EventsParams) =>
  http.get<Event[]>('/events', {
    params,
  })

export const get = (id: string) =>
  http.get<Event>('/events/get', {
    params: {
      id,
    },
  })
