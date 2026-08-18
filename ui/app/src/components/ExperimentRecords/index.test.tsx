/*
 * Copyright 2026 Chaos Mesh Authors.
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
import { render, screen } from '@/test-utils'

import ExperimentRecords from '.'

describe('<ExperimentRecords />', () => {
  it('renders nothing without target records', () => {
    const { container } = render(<ExperimentRecords status={{ experiment: {} }} />)

    expect(container).toBeEmptyDOMElement()
  })

  it('renders target state, counters, and event history', () => {
    render(
      <ExperimentRecords
        status={{
          experiment: {
            containerRecords: [
              {
                id: 'default/api/api-container',
                selectorKey: '.',
                phase: 'Injected',
                injectedCount: 2,
                recoveredCount: 1,
                events: [
                  {
                    type: 'Failed',
                    operation: 'Recover',
                    message: 'sandbox unavailable',
                    timestamp: '2026-08-18T12:00:00Z',
                  },
                ],
              },
            ],
          },
        }}
      />,
    )

    expect(screen.getByText('default/api/api-container')).toBeInTheDocument()
    expect(screen.getByText('Injected')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
    expect(screen.getByText('1')).toBeInTheDocument()
    expect(screen.getByText('1 events')).toBeInTheDocument()
    expect(screen.getByText('Recover Failed')).toBeInTheDocument()
    expect(screen.getByText('sandbox unavailable')).toBeInTheDocument()
  })
})
