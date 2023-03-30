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
import { act, render, screen } from 'test-utils'

import LoadFrom from '.'

jest.mock('lib/idb', () => {
  return {
    getDB: jest.fn().mockResolvedValue({
      getAll: jest.fn().mockResolvedValue([]),
    }),
  }
})

describe('<LoadFrom />', () => {
  test('loads and displays experiments and archives', async () => {
    await act(async () => {
      render(<LoadFrom />)
    })

    expect(screen.getByText('Experiments')).toBeInTheDocument()
    expect(screen.queryByText('No experiments found')).not.toBeInTheDocument()
  })

  test('loads and displays schedules', async () => {
    await act(async () => {
      render(<LoadFrom inSchedule />)
    })

    expect(screen.getByText('Schedules')).toBeInTheDocument()
    expect(screen.queryByText('No schedules found')).not.toBeInTheDocument()
  })
})
