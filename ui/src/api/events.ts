import { Event, EventsParams, GetEvents } from './events.type'

import { AxiosResponse } from 'axios'
import http from './http'

export const events: GetEvents = (params) =>
  http.get('/events', {
    params,
  })

// Without pods field
export const dryEvents: (params?: EventsParams) => Promise<AxiosResponse<Event[]>> = (params) =>
  http.get('/events/dry', {
    params,
  })
