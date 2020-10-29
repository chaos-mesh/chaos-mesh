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
  pods: Record<string, string[]> | string[]
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
  action: 'partition' | 'loss' | 'delay' | 'duplicate' | 'corrupt' | 'bandwidth' | ''
  bandwidth: ExperimentTargetNetworkBandwidth
  corrupt: ExperimentTargetNetworkCorrupt
  delay: ExperimentTargetNetworkDelay
  duplicate: ExperimentTargetNetworkDuplicate
  loss: ExperimentTargetNetworkLoss
  direction: 'from' | 'to' | 'both' | ''
  target?: ExperimentScope
}

export interface ExperimentTargetIO {
  action: 'latency' | 'fault' | 'attrOverride' | ''
  delay: string | undefined
  errno: number | undefined
  attr: object | string[]
  methods: string[]
  path: string
  percent: number
  volume_path: string
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
  fail_kern_request: FailKernelReq
}

export interface ExperimentTargetTime {
  clock_ids: string[]
  container_names: string[]
  time_offset: string
}

export interface ExperimentTargetStress {
  stressng_stressors: string
  stressors: {
    cpu: {
      workers: number
      load: number
      options: string[]
    } | null
    memory: {
      workers: number
      options: string[]
    } | null
  }
  container_name: string
}

export type ExperimentKind = 'PodChaos' | 'NetworkChaos' | 'IoChaos' | 'KernelChaos' | 'TimeChaos' | 'StressChaos'

export interface ExperimentTarget {
  kind: ExperimentKind | ''
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
