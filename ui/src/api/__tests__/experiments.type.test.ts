import ExperimentTI from 'api/experiments.type-ti'
import { createCheckers } from 'ts-interface-checker'

const { Experiment } = createCheckers(ExperimentTI)

const dummyExperiment = {
  uid: 'xxx',
  kind: 'PodChaos',
  namespace: 'default',
  name: 'pod-kill',
  created: 'xxx',
  status: 'Running',
}

describe('Check experiments type', () => {
  it('Experiment', () => {
    // Normal
    Experiment.check(dummyExperiment)
    Experiment.check({ ...dummyExperiment, status: 'Waiting' })

    // Abnormal
    expect(() => Experiment.check({ ...dummyExperiment, kind: 'HelloWorldChaos' })).toThrow(
      'value.kind is not a ExperimentKind'
    )
    expect(() => Experiment.check({ ...dummyExperiment, namespace: null })).toThrow('value.namespace is not a string')
    expect(() => Experiment.check({ ...dummyExperiment, name: null })).toThrow('value.name is not a string')
    expect(() => Experiment.check({ ...dummyExperiment, created: null })).toThrow('value.created is not a string')
    expect(() => Experiment.check({ ...dummyExperiment, status: 'Unknown' })).toThrow(
      'value.status is none of "Running", "Waiting", "Paused", "Failed", "Finished"'
    )
  })
})
