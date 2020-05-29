import http from './http'

export const state = () => http.get('/experiments/state')
