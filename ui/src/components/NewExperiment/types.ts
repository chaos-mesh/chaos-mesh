import { FormikContextType } from 'formik'

export interface ExperimentBasic {
  name: string
  namespace: string
}

export interface ExperimentScope {
  namespace_selectors: string[]
  label_selectors: object | string[]
  annotation_selectors: object | string[]
  phase_selectors: string[]
  mode: string
  value: string
}

export interface ExperimentTargetPod {
  action: string
  container_name?: string
}

export interface ExperimentTargetNetworkBandwidth {
  buffer: number
  limit: number
  minburst: number
  peakrate: number
  rate: string
}

export interface ExperimentTargetNetworkCorrupt {
  correlation: string
  corrupt: string
}

export interface ExperimentTargetNetworkDelay {
  latency: string
  correlation: string
  jitter: string
}

export interface ExperimentTargetNetworkDuplicate {
  correlation: string
  duplicate: string
}

export interface ExperimentTargetNetworkLoss {
  correlation: string
  loss: string
}

export interface ExperimentTargetNetwork {
  action: string
  bandwidth: ExperimentTargetNetworkBandwidth
  corrupt: ExperimentTargetNetworkCorrupt
  delay: ExperimentTargetNetworkDelay
  duplicate: ExperimentTargetNetworkDuplicate
  loss: ExperimentTargetNetworkLoss
}

export interface ExperimentTarget {
  kind: string
  pod_chaos: ExperimentTargetPod
  network_chaos: ExperimentTargetNetwork
  io_chaos?: any
  kernel_chaos?: any
  time_chaos?: any
  stress_chaos?: any
}

export interface ExperimentSchedule {
  cron: string
  duration: string
}

export interface Experiment extends ExperimentBasic {
  scope: ExperimentScope
  target: ExperimentTarget
  scheduler: ExperimentSchedule
}

export interface StepperState {
  activeStep: number
}

export enum StepperType {
  NEXT = 'NEXT',
  BACK = 'BACK',
  JUMP = 'JUMP',
  RESET = 'RESET',
}

export type StepperAction = {
  type: StepperType
  payload?: number
}

type StepperDispatch = (action: StepperAction) => void

export interface StepperContextProps {
  state: StepperState
  dispatch: StepperDispatch
}

export type FormikCtx = FormikContextType<Experiment>

export type StepperFormTargetProps = FormikCtx & {
  handleActionChange: (e: React.ChangeEvent<any>) => void
}
