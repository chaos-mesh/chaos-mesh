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
import http from './http'

let tokenInterceptorId: number

interface GCPToken {
  accessToken: string
  expiry: string
}

export const applyAPIAuthentication = (token: string | GCPToken) => {
  if (tokenInterceptorId !== undefined) {
    http.interceptors.request.eject(tokenInterceptorId)
  }

  const headers: {
    Authorization?: string
    'X-Authorization-Method'?: string
    'X-Authorization-AccessToken'?: string
    'X-Authorization-Expiry'?: string
  } =
    typeof token === 'string'
      ? {
          Authorization: `Bearer ${token}`,
        }
      : {
          'X-Authorization-Method': 'gcp',
          'X-Authorization-AccessToken': token.accessToken,
          'X-Authorization-Expiry': token.expiry,
        }

  tokenInterceptorId = http.interceptors.request.use((config) => {
    config.headers = {
      ...config.headers,
      ...headers,
    }

    return config
  })
}

export const resetAPIAuthentication = () => http.interceptors.request.eject(tokenInterceptorId)

let nsInterceptorId: number

export const applyNSParam = (ns: string) => {
  if (nsInterceptorId !== undefined) {
    http.interceptors.request.eject(nsInterceptorId)
  }

  nsInterceptorId = http.interceptors.request.use((config) => {
    if (/state|workflows|schedules|experiments|events|archives$/g.test(config.url!)) {
      config.params = {
        ...config.params,
        namespace: ns === 'All' ? null : ns,
      }
    }

    return config
  })
}
