import http from './http'

export const namespaces = () => http.get('/common/namespaces')
