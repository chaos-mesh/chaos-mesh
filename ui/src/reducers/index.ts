import { combineReducers } from 'redux'
import globalStatus from 'slices/globalStatus'
import navigation from 'slices/navigation'

export default combineReducers({
  navigation,
  globalStatus,
})
