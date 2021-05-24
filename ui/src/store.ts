import { AnyAction, configureStore, getDefaultMiddleware } from '@reduxjs/toolkit'
import { TypedUseSelectorHook, useDispatch, useSelector } from 'react-redux'

import rootReducer from 'reducers'

const middlewares = getDefaultMiddleware({
  serializableCheck: false, // warn: in order to use the global ConfirmDialog, disable the serializableCheck check
})

const genStore = () => {
  if (process.env.NODE_ENV === 'development') {
    const { createLogger } = require('redux-logger')

    const logger = createLogger({
      predicate: (_: any, action: AnyAction) => {
        if (action.type.includes('pending')) {
          return false
        }

        return true
      },
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
export const useStoreSelector: TypedUseSelectorHook<RootState> = useSelector
type StoreDispatch = typeof store.dispatch
export const useStoreDispatch = () => useDispatch<StoreDispatch>()

const store = genStore()

export default store
