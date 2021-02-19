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
