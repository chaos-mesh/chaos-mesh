import { FormikContextType } from 'formik'

export interface ExperimentBasic {
  name: string
  namespace: string
  labels: object | string[]
  annotations: object | string[]
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
  action: 'pod-failure' | 'pod-kill' | 'container-kill' | ''
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
  action: 'loss' | 'delay' | 'duplicate' | 'corrupt' | 'bandwidth' | ''
  bandwidth: ExperimentTargetNetworkBandwidth
  corrupt: ExperimentTargetNetworkCorrupt
  delay: ExperimentTargetNetworkDelay
  duplicate: ExperimentTargetNetworkDuplicate
  loss: ExperimentTargetNetworkLoss
}

export interface ExperimentTargetIO {
  action: 'delay' | 'errno' | 'mixed' | ''
  addr: string
  delay: string
  errno: string
  methods: string[]
  path: string
  percent: string
}

export interface CallchainFrame {
  funcname: string
  parameters: string
  predicate: string
}

export interface FailKernelReq {
  callchain: CallchainFrame[]
  failtype: number
  headers: string[]
  probability: number
  times: number
}

export interface ExperimentTargetKernel {
  fail_kernel_req: FailKernelReq
}

export interface ExperimentTargetTime {
  clock_ids: string[]
  container_names: string[]
  offset: string
}

export interface ExperimentTargetStress {
  stressng_stressors: string
  stressors: {
    cpu: {
      load: number
      options: string[]
      workers: number
    }
    memory: {
      options: string[]
      size: string
      workers: number
    }
  }
}

export interface ExperimentTarget {
  kind: string
  pod_chaos: ExperimentTargetPod
  network_chaos: ExperimentTargetNetwork
  io_chaos: ExperimentTargetIO
  kernel_chaos: ExperimentTargetKernel
  time_chaos: ExperimentTargetTime
  stress_chaos: ExperimentTargetStress
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
  handleActionChange?: (e: React.ChangeEvent<any>) => void
}
