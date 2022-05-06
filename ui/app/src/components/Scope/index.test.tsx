/*
 * Copyright 2022 Chaos Mesh Authors.
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
import { Formik } from 'formik'
import { fireEvent, render, screen, waitFor, within } from 'test-utils'

import Scope from '.'
import { scopeInitialValues } from '../AutoForm/data'

jest.mock('api', () => ({
  common: {
    labels: jest.fn().mockImplementation((ns) => {
      return Promise.resolve({ data: ns[0] === 'ns1' ? { app: 'ns1' } : { app: 'ns2' } })
    }),
    pods: jest.fn().mockImplementation(({ namespaces: ns, labelSelectors }) => {
      if (Object.keys(labelSelectors).length === 0) {
        return Promise.resolve({
          data: [
            {
              ip: '172.17.0.9',
              name: 'hello-chaos-mesh',
              namespace: 'ns1',
              state: 'Running',
            },
          ],
        })
      } else {
        return Promise.resolve({
          data:
            ns[0] === 'ns1'
              ? [
                  {
                    ip: '172.17.0.10',
                    name: 'app-1',
                    namespace: 'ns1',
                    state: 'Running',
                  },
                ]
              : [
                  {
                    ip: '172.17.0.11',
                    name: 'app-2',
                    namespace: 'ns2',
                    state: 'Running',
                  },
                ],
        })
      }
    }),
  },
}))

describe('<Scope />', () => {
  it('loads with the specified namespaces', () => {
    render(
      <Formik initialValues={scopeInitialValues} onSubmit={() => {}}>
        <Scope namespaces={['ns1', 'ns2']} />
      </Formik>
    )

    expect(screen.getByText('No pods found')).toBeInTheDocument()
  })

  it('loads and then choose a namespace', async () => {
    render(
      <Formik initialValues={scopeInitialValues} onSubmit={() => {}}>
        <Scope namespaces={['ns1', 'ns2']} />
      </Formik>
    )

    const nsSelectors = screen.getByRole('combobox', { name: 'Namespace Selectors' })
    nsSelectors.focus()

    fireEvent.keyDown(nsSelectors, { key: 'ArrowDown' })
    fireEvent.keyDown(nsSelectors, { key: 'ArrowDown' }) // Move to the first option.
    fireEvent.keyDown(nsSelectors, { key: 'Enter' })

    await waitFor(() => {
      expect(screen.getByText('hello-chaos-mesh')).toBeInTheDocument()
    })
  })

  it('loads and then choose a namespace and a label', async () => {
    render(
      <Formik initialValues={scopeInitialValues} onSubmit={() => {}}>
        <Scope namespaces={['ns1', 'ns2']} />
      </Formik>
    )

    const nsSelectors = screen.getByRole('combobox', { name: 'Namespace Selectors' })
    nsSelectors.focus()

    fireEvent.keyDown(nsSelectors, { key: 'ArrowDown' })
    fireEvent.keyDown(nsSelectors, { key: 'ArrowDown' }) // Move to the first option.
    fireEvent.keyDown(nsSelectors, { key: 'Enter' })

    const nsLabels = screen.getByRole('combobox', { name: 'Label Selectors' })
    nsLabels.focus()

    fireEvent.keyDown(nsLabels, { key: 'ArrowDown' })

    expect(within(screen.getByRole('listbox')).getByText('app: ns1')).toBeInTheDocument()

    fireEvent.keyDown(nsLabels, { key: 'ArrowDown' })
    fireEvent.keyDown(nsLabels, { key: 'Enter' })

    await waitFor(() => {
      expect(screen.getByText('172.17.0.10')).toBeInTheDocument()
      expect(screen.getByText('app-1')).toBeInTheDocument()
    })
  })
})
