import { RootState, useStoreDispatch } from 'store'

import http from 'api/http'
import { setTokenIntercepterNumber } from 'slices/globalStatus'
import { useSelector } from 'react-redux'

export const useTokenHandler = () => {
  const dispatch = useStoreDispatch()
  const tokenIntercepterNumber = useSelector((state: RootState) => state.globalStatus.tokenIntercepterNumber)

  const tokenSubmitHandler = (token: string) => {
    if (tokenIntercepterNumber !== -1) {
      http.interceptors.request.eject(tokenIntercepterNumber)
    }
    const newTokenIntercepterNumber = http.interceptors.request.use((config) => {
      config.headers = {
        Authorization: `Bearer ${token}`,
      }
      return config
    })
    dispatch(setTokenIntercepterNumber(newTokenIntercepterNumber))
    window.localStorage.setItem('chaos-mesh-token', token)
  }

  return tokenSubmitHandler
}
