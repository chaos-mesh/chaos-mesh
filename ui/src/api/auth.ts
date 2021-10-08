import http from 'api/http'

let tokenInterceptorId: number

interface GCPToken {
  accessToken: string
  expiry: string
}

export const token = (token: string | GCPToken) => {
  if (tokenInterceptorId !== undefined) {
    http.interceptors.request.eject(tokenInterceptorId)
  }

  const headers =
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

export const resetToken = () => http.interceptors.request.eject(tokenInterceptorId)

let nsInterceptorId: number

export const namespace = (ns: string) => {
  if (nsInterceptorId !== undefined) {
    http.interceptors.request.eject(nsInterceptorId)
  }

  nsInterceptorId = http.interceptors.request.use((config) => {
    if (/state|workflows|schedules|experiments|events(\/dry)?|archives$/g.test(config.url!)) {
      config.params = {
        ...config.params,
        namespace: ns === 'All' ? null : ns,
      }
    }

    return config
  })
}
