import { PayloadAction } from '@reduxjs/toolkit'
import { StateOfExperiments } from 'api/experiments.type'

export type ExperimentsAction = PayloadAction<StateOfExperiments | string[] | { [key: string]: string[] }>
