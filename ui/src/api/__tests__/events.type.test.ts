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
import EventTI from 'api/events.type-ti'
import { createCheckers } from 'ts-interface-checker'

const { Event, EventPod } = createCheckers(EventTI)

const dummyEvent = {
  id: 1,
  experiment_id: 'xxx',
  experiment: 'xxx',
  namespace: 'default',
  kind: 'PodChaos',
  message: 'xxx',
  start_time: 'xxx',
  finish_time: 'xxx',
}

const dummyPod = {
  id: 1,
  pod_ip: 'xxx',
  pod_name: 'xxx',
  namespace: 'xxx',
  action: 'xxx',
  message: 'xxx',
}

describe('Check events type', () => {
  it('Event', () => {
    // Normal
    Event.check(dummyEvent)

    // Abnormal
    expect(() => Event.check({ ...dummyEvent, experiment: null })).toThrow('value.experiment is not a string')
    expect(() => Event.check({ ...dummyEvent, namespace: null })).toThrow('value.namespace is not a string')
    expect(() => Event.check({ ...dummyEvent, kind: 'HelloWorldChaos' })).toThrow('value.kind is not a ExperimentKind')
    expect(() => Event.check({ ...dummyEvent, message: null })).toThrow('value.message is not a string')
    expect(() => Event.check({ ...dummyEvent, start_time: null })).toThrow('value.start_time is not a string')
    expect(() => Event.check({ ...dummyEvent, finish_time: null })).toThrow('value.finish_time is not a string')
  })

  it('EventPod', () => {
    // Normal
    EventPod.strictCheck(dummyPod)
  })
})
