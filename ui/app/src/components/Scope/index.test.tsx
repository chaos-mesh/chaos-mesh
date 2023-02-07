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
    commonLabelsGet: jest.fn().mockImplementation(({ podNamespaceList: nsStr }) => {
      const ns = nsStr.split(',')

      return Promise.resolve({ data: ns[0] === 'ns1' ? { app: 'ns1' } : { app: 'ns2' } })
    }),
    commonPodsPost: jest.fn().mockImplementation(({ request: { namespaces: ns, labelSelectors } }) => {
      const app1 = {
        ip: '172.17.0.10',
        name: 'app-1',
        namespace: 'ns1',
        state: 'Running',
      }
      const app2 = {
        ip: '172.17.0.11',
        name: 'app-2',
        namespace: 'ns2',
        state: 'Running',
      }
      if (Object.keys(labelSelectors).length === 0) {
        return Promise.resolve({
          data:
            ns[0] === 'ns1'
              ? [
                  {
                    ip: '172.17.0.9',
                    name: 'hello-chaos-mesh',
                    namespace: 'ns1',
                    state: 'Running',
                  },
                  app1,
                ]
              : [app2],
        })
      } else {
        return Promise.resolve({
          data: ns[0] === 'ns1' ? [app1] : [app2],
        })
      }
    }),
  },
}))

const Default = () => (
  <Formik initialValues={scopeInitialValues({ hasSelector: true })} onSubmit={() => {}}>
    <Scope env="k8s" kind="PodChaos" namespaces={['ns1', 'ns2']} />
  </Formik>
)

describe('<Scope />', () => {
  it('disables when kind is AWSChaos', () => {
    render(
      <Formik initialValues={scopeInitialValues({ hasSelector: true })} onSubmit={() => {}}>
        <Scope env="k8s" kind="AWSChaos" namespaces={['ns1', 'ns2']} />
      </Formik>
    )

    expect(screen.getByText('AWSChaos does not need to define the scope.')).toBeInTheDocument()
  })

  it('first load', () => {
    render(<Default />)

    expect(screen.getByText('No Pods found.')).toBeInTheDocument()
  })

  it('loads and then choose a namespace', async () => {
    render(<Default />)

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
    render(<Default />)

    const nsSelectors = screen.getByRole('combobox', { name: 'Namespace Selectors' })
    nsSelectors.focus()

    fireEvent.keyDown(nsSelectors, { key: 'ArrowDown' })
    fireEvent.keyDown(nsSelectors, { key: 'ArrowDown' })
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

  it('reselects namespace', async () => {
    render(<Default />)

    const nsSelectors = screen.getByRole('combobox', { name: 'Namespace Selectors' })
    nsSelectors.focus()

    fireEvent.keyDown(nsSelectors, { key: 'ArrowDown' })
    fireEvent.keyDown(nsSelectors, { key: 'ArrowDown' })
    fireEvent.keyDown(nsSelectors, { key: 'Enter' })
    fireEvent.keyDown(nsSelectors, { key: 'Backspace' }) // Delete n1.
    fireEvent.keyDown(nsSelectors, { key: 'ArrowDown' })
    fireEvent.keyDown(nsSelectors, { key: 'ArrowDown' })
    fireEvent.keyDown(nsSelectors, { key: 'ArrowDown' }) // Select n2.
    fireEvent.keyDown(nsSelectors, { key: 'Enter' })

    await waitFor(() => {
      expect(screen.getByText('172.17.0.11')).toBeInTheDocument()
      expect(screen.getByText('app-2')).toBeInTheDocument()
    })
  })
})
