import { AxiosResponse } from 'axios'
import { Event } from './events.type'
import http from './http'

export const events: () => Promise<AxiosResponse<Event[]>> = () => http.get('/events')
