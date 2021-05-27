import { combineReducers } from 'redux'
import experiments from 'slices/experiments'
import globalStatus from 'slices/globalStatus'
import navigation from 'slices/navigation'
import settings from 'slices/settings'
import workflows from 'slices/workflows'

export default combineReducers({
  settings,
  globalStatus,
  navigation,
  experiments,
  workflows,
})
