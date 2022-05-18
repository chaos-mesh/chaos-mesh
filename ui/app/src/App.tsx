import { DndProvider } from 'react-dnd'
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
import type { FC } from 'react'
import { HTML5Backend } from 'react-dnd-html5-backend'
import IntlProvider from './IntlProvider'
import { Provider as StoreProvider } from 'react-redux'
import ThemeProvider from './ThemeProvider'
import TopContainer from 'components/TopContainer'
import store from './store'

interface AppProps {
  forTesting?: boolean
}

const App: FC<AppProps> = ({ forTesting, children }) => {
  const rendered = children || <TopContainer />
  const RealWorldOnlyProviders: FC = ({ children }) => <DndProvider backend={HTML5Backend}>{children}</DndProvider>

  return (
    <StoreProvider store={store}>
      <ThemeProvider>
        <IntlProvider>
          {!forTesting ? <RealWorldOnlyProviders>{rendered}</RealWorldOnlyProviders> : rendered}
        </IntlProvider>
      </ThemeProvider>
    </StoreProvider>
  )
}

export default App
