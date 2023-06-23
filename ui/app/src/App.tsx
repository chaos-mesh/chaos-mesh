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
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import type { FC } from 'react'
import { DndProvider } from 'react-dnd'
import { HTML5Backend } from 'react-dnd-html5-backend'
import { Provider as StoreProvider } from 'react-redux'
import { RouterProvider } from 'react-router-dom'

import store from './store'

import IntlProvider from './IntlProvider'
import ThemeProvider from './ThemeProvider'
import router from './router'

const queryClient = new QueryClient()

interface AppProps {
  forTesting?: boolean
}

const App: FC<AppProps> = ({ forTesting, children }) => {
  const rendered = children || <RouterProvider router={router} />
  const RealWorldOnlyProviders: FC = ({ children }) => <DndProvider backend={HTML5Backend}>{children}</DndProvider>

  return (
    <StoreProvider store={store}>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider>
          <IntlProvider>
            {!forTesting ? <RealWorldOnlyProviders>{rendered}</RealWorldOnlyProviders> : rendered}
          </IntlProvider>
        </ThemeProvider>
        <ReactQueryDevtools initialIsOpen={false} />
      </QueryClientProvider>
    </StoreProvider>
  )
}

export default App
