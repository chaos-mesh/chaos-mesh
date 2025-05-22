/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import { configureStore, getDefaultMiddleware } from '@reduxjs/toolkit'
import { type TypedUseSelectorHook, useDispatch, useSelector } from 'react-redux'

import rootReducer from './reducers'

const middlewares = getDefaultMiddleware({
  serializableCheck: false, // warn: in order to use the global ConfirmDialog, disable the serializableCheck check
})

const genStore = () => {
  const store = configureStore({
    reducer: rootReducer,
    middleware: middlewares,
    devTools: import.meta.env.DEV,
  })

  return store
}

export type RootState = ReturnType<typeof rootReducer>
export const useStoreSelector: TypedUseSelectorHook<RootState> = useSelector
type StoreDispatch = typeof store.dispatch
export const useStoreDispatch = () => useDispatch<StoreDispatch>()

const store = genStore()

export default store
