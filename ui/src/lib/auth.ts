import { RootState, useStoreDispatch } from 'store'

import http from 'api/http'
import { setTokenInterceptor } from 'slices/globalStatus'
import { useSelector } from 'react-redux'

export const useToken = () => {
  const interceptor = useSelector((state: RootState) => state.globalStatus.tokenInterceptor)
  const dispatch = useStoreDispatch()

  const register = (token: string) => {
    if (interceptor !== -1) {
      http.interceptors.request.eject(interceptor)
    }

    const newInterceptor = http.interceptors.request.use((config) => {
      config.headers = {
        ...config.headers,
        Authorization: `Bearer ${token}`,
      }

      return config
    })

    dispatch(setTokenInterceptor(newInterceptor))
  }

  return register
}
