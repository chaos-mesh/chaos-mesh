import { PayloadAction } from '@reduxjs/toolkit'

export interface StateOfExperiments {
  total: number
  running: number
  paused: number
  failed: number
  finished: number
}

export type GlobalStatusAction = PayloadAction<StateOfExperiments>
