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

import { RenderOptions, render } from '@testing-library/react'

import App from './App'
import { IntlProvider } from 'react-intl'
import { ReactElement } from 'react'
import flat from 'flat'
import messages from 'i18n/messages'

const AllTheProviders: React.FC = ({ children }) => (
  <App>
    <IntlProvider messages={flat(messages['en'])} locale="en" defaultLocale="en">
      {children}
    </IntlProvider>
  </App>
)

const customRender = (ui: ReactElement, options?: Omit<RenderOptions, 'wrapper'>) =>
  render(ui, { wrapper: AllTheProviders, ...options })

export * from '@testing-library/react'
export { customRender as render }
