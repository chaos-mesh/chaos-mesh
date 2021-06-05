import { Experiment } from './experiments.type'

export type Schedule = { is: 'schedule' } & Omit<Experiment, 'is' | 'status'>
