import axios, { AxiosError } from 'axios'

import { setAlert } from 'slices/globalStatus'
import store from 'store'

interface ErrorData {
  code: number
  type: string
  message: string
  full_text: string
}

const http = axios.create({
  baseURL: '/api',
})

http.interceptors.response.use(undefined, (error: AxiosError<ErrorData>) => {
  const data = error.response?.data

  if (data) {
    // slice(10): error.api.xxx => xxx
    const type = data.type.slice(10)

    switch (type) {
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
