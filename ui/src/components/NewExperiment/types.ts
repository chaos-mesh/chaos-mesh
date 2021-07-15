export interface Metadata {
  name: string
  namespace: string
  labels?: string[]
  annotations?: string[]
}

interface Selector {
  namespaces: string[]
  labelSelectors?: string[]
  annotationSelectors?: string[]
  phaseSelectors?: string[]
  pods?: string[]
}

export interface Scope {
  selector: Selector
  mode: string
  value?: string
}

export interface Pod {
  action: 'pod-failure' | 'pod-kill' | 'container-kill'
  containerNames?: string[]
}

export interface NetworkLoss {
  loss: string
  correlation: string
}

export interface NetworkDelay {
  latency: string
  jitter: string
  correlation: string
}

export interface NetworkDuplicate {
  duplicate: string
  correlation: string
}

export interface NetworkCorrupt {
  corrupt: string
  correlation: string
}

export interface NetworkBandwidth {
  rate: string
  limit: number
  buffer: number
  minburst: number
  peakrate: number
}

export interface Network {
  action: 'partition' | 'loss' | 'delay' | 'duplicate' | 'corrupt' | 'bandwidth'
  loss?: NetworkLoss
  delay?: NetworkDelay
  duplicate?: NetworkDuplicate
  corrupt?: NetworkCorrupt
  bandwidth?: NetworkBandwidth
  direction?: 'from' | 'to' | 'both'
  target?: Selector
}

export interface IO {
  action: 'latency' | 'fault' | 'attrOverride'
  delay?: string
  errno?: number
  attr?: object | string[]
  volumePath: string
  path: string
  percent: number
  methods: string[]
}

export interface Stress {
  stressors: {
    cpu?: {
      workers: number
      load: number
      options: string[]
    }
    memory?: {
      workers: number
      size: string
      options: string[]
    }
  }
  stressngStressors: string
  containerNames: string
}

export interface Frame {
  funcname: string
  parameters: string
  predicate: string
}

export interface FailKernelReq {
  callchain: Frame[]
  failtype: number
  headers: string[]
  probability: number
  times: number
}

export interface Kernel {
  failKernRequest: FailKernelReq
}

export interface Time {
  timeOffset: string
  clockIds: string[]
  containerNames: string[]
}

export interface DNS {
  action: 'error' | 'random'
  patterns: string[]
  containerNames?: string[]
}

export interface ExperimentTarget {
  PodChaos: Pod
  NetworkChaos: Network
  IOChaos: IO
  StressChaos: Stress
  KernelChaos: Kernel
  TimeChaos: Time
  DNSChaos: DNS
  AWSChaos: DNS
  GCPChaos: DNS
}

export type ExperimentKind = keyof ExperimentTarget

export interface Experiment<K extends ExperimentKind = any> {
  metadata: Metadata
  spec: Scope &
    Pick<ExperimentTarget, K>[K] & {
      duration?: string
    }
}
