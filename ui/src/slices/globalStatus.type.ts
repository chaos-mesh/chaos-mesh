import { PayloadAction } from '@reduxjs/toolkit'
import { StateOfExperiments } from 'api/experiments.type'

export type GlobalStatusAction = PayloadAction<StateOfExperiments>
