/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
  podPhaseSelectors?: string[]
  pods?: string[]
  physicalMachines?: string[]
}

export interface Scope {
  selector: Selector
  mode: string
  value?: string
}

export interface AWS {
  action: 'ec2-stop' | 'ec2-restart' | 'detach-volume'
  secretName: string
  awsRegion: string
  ec2Instance: string
  volumeID?: string
  deviceName?: string
}

export interface DNS {
  action: 'error' | 'random'
  patterns: string[]
  containerNames?: string[]
}

export interface GCP {
  action: 'node-stop' | 'node-reset' | 'disk-loss'
  secretName: string
  project: string
  zone: string
  instance: string
  deviceNames?: string[]
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

export interface Pod {
  action: 'pod-failure' | 'pod-kill' | 'container-kill'
  containerNames?: string[]
}

export interface ResourceScale {
  namespace: string
  name: string
  resourceType: 'daemonset' | 'deployment' | 'replicaset' | 'statefulset'
  applyReplicas: number
  recoverReplicas: number
}

export interface RollingRestart {
  namespace: string
  name: string
  type: 'daemonset' | 'deployment' | 'statefulset'
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

export interface Time {
  timeOffset: string
  clockIds: string[]
  containerNames: string[]
}

export interface ExperimentType {
  AWSChaos: AWS
  AzureChaos?: unknown
  CiliumChaos?: unknown
  CloudStackVMChaos?: unknown
  DNSChaos: DNS
  GCPChaos: GCP
  HTTPChaos?: unknown
  IOChaos: IO
  JVMChaos?: unknown
  K8SChaos?: unknown
  KernelChaos: Kernel
  NetworkChaos: Network
  PodChaos: Pod
  ResourceScaleChaos: ResourceScale
  RollingRestartChaos: RollingRestart
  StressChaos: Stress
  TimeChaos: Time
  PhysicalMachineChaos?: unknown
  BlockChaos?: unknown
}

export type ExperimentKind = keyof ExperimentType

export interface Experiment<K extends ExperimentKind> {
  metadata: Metadata
  spec: Scope &
    ExperimentType[K] & {
      duration?: string
    }
}
