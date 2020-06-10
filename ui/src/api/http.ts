import axios, { AxiosError } from 'axios'

import { setAlert, setAlertOpen } from 'slices/globalStatus'
import store from 'store'

const http = axios.create({
  baseURL: '/api',
})

http.interceptors.response.use(undefined, (error: AxiosError) => {
  if (error.response?.config.url === '/experiments/state') {
    return
  }

  const data = error.response?.data

  if (data) {
    store.dispatch(
      setAlert({
        type: 'error',
        message: data.message,
      })
    )
    store.dispatch(setAlertOpen(true))
  }

  return Promise.reject(error)
})

export default http
