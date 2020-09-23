import { combineReducers } from 'redux'
import experiments from 'slices/experiments'
import globalStatus from 'slices/globalStatus'
import navigation from 'slices/navigation'
import settings from 'slices/settings'

export default combineReducers({
  settings,
  globalStatus,
  navigation,
  experiments,
})
