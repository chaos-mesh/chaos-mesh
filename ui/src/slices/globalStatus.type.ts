import { PayloadAction } from '@reduxjs/toolkit'
import { StateOfExperiments } from 'api/experiments.type'

export interface Alert {
  type: 'success' | 'error'
  message: string
}

export type GlobalStatusAction = PayloadAction<StateOfExperiments>
