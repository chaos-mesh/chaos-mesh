import axios, { AxiosError } from 'axios'

import { setAlert } from 'slices/globalStatus'
import store from 'store'

interface ErrorData {
  status: 'error'
  code: string
  message: string
  full_text: string
}

const http = axios.create({
  baseURL: '/api',
})

http.interceptors.response.use(undefined, (error: AxiosError<ErrorData>) => {
  const data = error.response?.data

  if (data) {
    // error.api.xxx => xxx
    switch (data.code.slice(10)) {
      case 'invalid_request':
        if (data.message.includes('Unauthorized')) {
          store.dispatch(
            setAlert({
              type: 'error',
              message: 'Please check the validity of the token',
            })
          )

          break
        }
      // eslint-disable-next-line
      case 'no_cluster_privilege':
      case 'no_namespace_privilege':
      default:
        store.dispatch(
          setAlert({
            type: 'error',
            message: data.message || 'An unknown error occurred',
          })
        )

        break
    }
  }

  return Promise.reject(error)
})

export default http
