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

import { act, fireEvent, render, screen } from 'test-utils'

import LoadFrom from '.'

const experiments = [
  {
    namespace: 'default',
    name: 'random-pod-failure',
    kind: 'PodChaos',
    uid: 'e626701f-beaa-47de-91ce-767b856d8bec',
    created_at: '2021-11-04T04:17:09Z',
    status: 'finished',
  },
]

const schedules = [
  {
    namespace: 'default',
    name: 'sch-random-pod-failure',
    kind: 'PodChaos',
    uid: '63932612-978e-4aa8-81f5-98964ea82f9c',
    created_at: '2021-11-12T03:44:15Z',
    status: 'running',
  },
]

const archives = [
  {
    namespace: 'default',
    name: 'network-delay-90ms',
    kind: 'NetworkChaos',
    uid: '5ccdfb1e-5b12-4d3c-b065-7669552cc0f7',
    created_at: '2021-11-11T08:11:15Z',
  },
]

const scheduleArchives = [
  {
    namespace: 'default',
    name: 'sch-network-delay-90ms',
    kind: 'Schedule',
    uid: 'af9daeb3-4cfc-48b6-a547-e1003810b628',
    created_at: '2021-11-12T04:05:35Z',
  },
]

jest.mock('api', () => {
  return {
    experiments: {
      experimentsGet: jest.fn().mockResolvedValue({ data: experiments }),
      experimentsUidGet: jest.fn().mockResolvedValue({ data: { kube_object: { spec: {} } } }),
    },
    schedules: {
      schedulesGet: jest.fn().mockResolvedValue({ data: schedules }),
    },
    archives: {
      archivesGet: jest.fn().mockResolvedValue({ data: archives }),
      archivesSchedulesGet: jest.fn().mockResolvedValue({ data: scheduleArchives }),
    },
  }
})

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

    screen.getByText('random-pod-failure')
  })

  test('props inSchedule', async () => {
    await act(async () => {
      render(<LoadFrom inSchedule />)
    })

    screen.getByText('sch-random-pod-failure')
    screen.getByText('sch-network-delay-90ms')
  })

  test('loads an experiment', async () => {
    function callback(data: any) {
      expect(data).not.toBeUndefined()
    }

    await act(async () => {
      render(<LoadFrom callback={callback} />)
    })

    fireEvent.click(screen.getByText('random-pod-failure'))
  })
})
