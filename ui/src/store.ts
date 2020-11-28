import { AnyAction, configureStore, getDefaultMiddleware } from '@reduxjs/toolkit'

import rootReducer from 'reducers'
import { useDispatch } from 'react-redux'

const middlewares = [...getDefaultMiddleware()]
const ignoreActions = [
  'experiments/state/pending',
  'common/chaos-available-namespaces/pending',
  'common/labels/pending',
  'common/annotations/pending',
  'common/pods/pending',
]

const genStore = () => {
  if (process.env.NODE_ENV === 'development') {
    const { createLogger } = require('redux-logger')

    const logger = createLogger({
      predicate: (_: any, action: AnyAction) => !ignoreActions.includes(action.type),
    })

    middlewares.push(logger)
  }

  const store = configureStore({
    reducer: rootReducer,
    middleware: middlewares,
    devTools: process.env.NODE_ENV !== 'production',
  })

  return store
}

export type RootState = ReturnType<typeof rootReducer>
type StoreDispatch = typeof store.dispatch
export const useStoreDispatch = () => useDispatch<StoreDispatch>()

const store = genStore()

export default store
