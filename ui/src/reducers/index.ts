import { combineReducers } from 'redux'
import experiments from 'slices/experiments'
import globalStatus from 'slices/globalStatus'
import navigation from 'slices/navigation'

export default combineReducers({
  navigation,
  globalStatus,
  experiments,
})
