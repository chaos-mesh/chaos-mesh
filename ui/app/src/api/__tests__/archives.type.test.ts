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
import ArchiveTI from 'api/archives.type-ti'
import { createCheckers } from 'ts-interface-checker'

const { Archive } = createCheckers(ArchiveTI)

const dummyArchive = {
  uid: 'xxx',
  kind: 'PodChaos',
  namespace: 'default',
  name: 'pod-kill',
  start_time: 'xxx',
  finish_time: 'xxx',
}

describe('Check archives type', () => {
  it('Archive', () => {
    // Normal
    Archive.check(dummyArchive)

    // Abnormal
    expect(() => Archive.check({ ...dummyArchive, kind: 'HelloWorldChaos' })).toThrow(
      'value.kind is not a ExperimentKind'
    )
    expect(() => Archive.check({ ...dummyArchive, namespace: null })).toThrow('value.namespace is not a string')
    expect(() => Archive.check({ ...dummyArchive, name: null })).toThrow('value.name is not a string')
    expect(() => Archive.check({ ...dummyArchive, start_time: null })).toThrow('value.start_time is not a string')
    expect(() => Archive.check({ ...dummyArchive, finish_time: null })).toThrow('value.finish_time is not a string')
  })
})
