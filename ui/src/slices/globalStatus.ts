import { Alert } from './globalStatus.type'
import { createSlice } from '@reduxjs/toolkit'

const initialState: {
  alert: Alert
  alertOpen: boolean
  searchModalOpen: boolean
  tokenInterceptorNumber: number
  namespaceInterceptorNumber: number
  hasPrivilege: boolean
  isValidToken: boolean
  isPrivilegedToken: boolean
} = {
  alert: {
    type: 'success',
    message: '',
  },
  alertOpen: false,
  searchModalOpen: false,
  tokenInterceptorNumber: -1,
  namespaceInterceptorNumber: -1,
  hasPrivilege: true,
  isValidToken: true,
  isPrivilegedToken: true,
}

const globalStatusSlice = createSlice({
  name: 'globalStatus',
  initialState,
  reducers: {
    setAlert(state, action) {
      state.alert = action.payload
    },
    setAlertOpen(state, action) {
      state.alertOpen = action.payload
    },
    setSearchModalOpen(state, action) {
      state.searchModalOpen = action.payload
    },
    setTokenInterceptorNumber(state, action) {
      state.tokenInterceptorNumber = action.payload
    },
    setNameSpaceInterceptorNumber(state, action) {
      state.namespaceInterceptorNumber = action.payload
    },
    setHasPrivilege(state, action) {
      state.hasPrivilege = action.payload
    },
    setIsValidToken(state, action) {
      state.isValidToken = action.payload
    },
    setIsPrivilegedToken(state, action) {
      state.isPrivilegedToken = action.payload
    },
  },
})

export const {
  setAlert,
  setAlertOpen,
  setSearchModalOpen,
  setTokenInterceptorNumber,
  setNameSpaceInterceptorNumber,
  setHasPrivilege,
  setIsValidToken,
  setIsPrivilegedToken,
} = globalStatusSlice.actions

export default globalStatusSlice.reducer
