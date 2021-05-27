import { Experiment } from './experiments.type'
import { ExperimentKind } from 'components/NewExperiment/types'

export type Schedule = Omit<Experiment, 'kind' | 'status' | 'events'> & { type: ExperimentKind }
