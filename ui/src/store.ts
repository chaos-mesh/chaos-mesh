import { configureStore, getDefaultMiddleware } from '@reduxjs/toolkit'

import { Middleware } from 'redux'
import rootReducer from 'reducers'
import { useDispatch } from 'react-redux'

const middlewares: Middleware[] = [...getDefaultMiddleware()]

const genStore = () => {
  if (process.env.NODE_ENV === 'development') {
    const { logger } = require('redux-logger')

    middlewares.push(logger)
  }

  const store = configureStore({
    reducer: rootReducer,
    middleware: middlewares,
  })

  return store
}

const store = genStore()

export type RootState = ReturnType<typeof rootReducer>

type StoreDispatch = typeof store.dispatch
export const useStoreDispatch = () => useDispatch<StoreDispatch>()

export default store
