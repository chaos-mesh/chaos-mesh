import axios, { AxiosError } from 'axios'
import { setAlert, setAlertOpen, setHasPrivilege, setIsPrivilegedToken, setIsValidToken } from 'slices/globalStatus'

import store from 'store'

const http = axios.create({
  baseURL: '/api',
})

http.interceptors.response.use(undefined, (error: AxiosError) => {
  const data = error.response?.data

  if (data) {
    if (data.code === 'error.api.no_cluster_privilege' || data.code === 'error.api.no_namespace_privilege') {
      store.dispatch(setHasPrivilege(false))
    } else if (data.code === 'error.api.internal_server_error' && data.message.includes('forbidden')) {
      store.dispatch(setIsPrivilegedToken(false))
    } else if (data.code === 'error.api.invalid_request' && data.message.includes('Unauthorized')) {
      store.dispatch(setIsValidToken(false))
    }

    store.dispatch(
      setAlert({
        type: 'error',
        message: data.message || 'An unknown error occurred. Please check your http request.',
      })
    )
    store.dispatch(setAlertOpen(true))
  }

  return Promise.reject(error)
})

export default http
