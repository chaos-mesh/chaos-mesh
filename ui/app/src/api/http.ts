/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import axios, { AxiosError, AxiosRequestConfig } from 'axios'

import store from 'store'

import { setAlert } from 'slices/globalStatus'

interface ErrorData {
  code: number
  type: string
  message: string
  full_text: string
}

const http = axios.create({ baseURL: '/api' })

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
      // eslint-disable-next-line no-fallthrough
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

export const customInstance = <T>(config: AxiosRequestConfig): Promise<T> => {
  const promise = http(config).then(({ data }) => data)

  return promise
}

export default http
