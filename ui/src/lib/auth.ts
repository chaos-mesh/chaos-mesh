import { RootState, useStoreDispatch } from 'store'
import { setNameSpaceInterceptorNumber, setTokenInterceptorNumber } from 'slices/globalStatus'

import { ActionCreatorWithPayload } from '@reduxjs/toolkit'
import { AxiosRequestConfig } from 'axios'
import http from 'api/http'
import { useSelector } from 'react-redux'

type InterceptorNumberStateName = 'tokenInterceptorNumber' | 'namespaceInterceptorNumber'

function useInterceptorRegistry(
  stateName: InterceptorNumberStateName,
  interceptor: (
    storageValue: string
  ) => (value: AxiosRequestConfig) => AxiosRequestConfig | Promise<AxiosRequestConfig>,
  action: ActionCreatorWithPayload<any, string>,
  storageKey: string
) {
  const dispatch = useStoreDispatch()
  const interceptorNumber = useSelector((state: RootState) => state.globalStatus[stateName])

  const registry = (storageValue: string) => {
    if (interceptorNumber !== -1) {
      http.interceptors.request.eject(interceptorNumber)
    }
    const newInterceptorNumber = http.interceptors.request.use(interceptor(storageValue))
    dispatch(action(newInterceptorNumber))
    window.localStorage.setItem(storageKey, storageValue)
  }

  return registry
}

export const useTokenRegistry = () => {
  return useInterceptorRegistry(
    'tokenInterceptorNumber',
    (token) => (config) => {
      config.headers = {
        Authorization: `Bearer ${token}`,
      }
      return config
    },
    setTokenInterceptorNumber,
    'chaos-mesh-token'
  )
}

export const useNameSpaceRegistry = () => {
  return useInterceptorRegistry(
    'namespaceInterceptorNumber',
    (namespace) => (config) => {
      console.log(namespace)
      config.params = {
        namespace,
      }
      return config
    },
    setNameSpaceInterceptorNumber,
    'chaos-mesh-namespace'
  )
}
