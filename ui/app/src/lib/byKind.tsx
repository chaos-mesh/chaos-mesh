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
import LinearScaleIcon from '@mui/icons-material/LinearScale'
import TimelapseIcon from '@mui/icons-material/Timelapse'
import { SvgIcon } from '@mui/material'

import { type ExperimentKind } from '@/components/NewExperiment/types'
import i18n from '@/components/T'

import AWSIcon from '@/images/chaos/aws.svg?react'
import DiskIcon from '@/images/chaos/disk.svg?react'
import DNSIcon from '@/images/chaos/dns.svg?react'
import GCPIcon from '@/images/chaos/gcp.svg?react'
import YCIcon from '@/images/chaos/yc.svg?react'
import HTTPIcon from '@/images/chaos/http.svg?react'
import FileSystemIOIcon from '@/images/chaos/io.svg?react'
import JavaIcon from '@/images/chaos/java.svg?react'
import LinuxKernelIcon from '@/images/chaos/kernel.svg?react'
import NetworkIcon from '@/images/chaos/network.svg?react'
import PodLifecycleIcon from '@/images/chaos/pod.svg?react'
import ProcessIcon from '@/images/chaos/process.svg?react'
import StressIcon from '@/images/chaos/stress.svg?react'
import ClockIcon from '@/images/chaos/time.svg?react'
import K8SIcon from '@/images/k8s.svg?react'
import PhysicIcon from '@/images/physic.svg?react'

export function iconByKind(kind: string, size: 'small' | 'inherit' | 'medium' | 'large' = 'medium') {
  let icon

  switch (kind) {
    case 'k8s':
      icon = <K8SIcon />
      break
    case 'physic':
    case 'PhysicalMachineChaos':
      icon = <PhysicIcon />
      break
    case 'AWSChaos':
      icon = <AWSIcon />
      break
    case 'BlockChaos':
      // TODO: add icon for BlockChaos
      icon = <FileSystemIOIcon />
      break
    case 'DiskChaos':
      icon = <DiskIcon />
      break
    case 'DNSChaos':
      icon = <DNSIcon />
      break
    case 'GCPChaos':
      icon = <GCPIcon />
      break
    case 'YCChaos':
      icon = <YCIcon />
      break
    case 'IOChaos':
      icon = <FileSystemIOIcon />
      break
    case 'HTTPChaos':
      icon = <HTTPIcon />
      break
    case 'JVMChaos':
      icon = <JavaIcon />
      break
    case 'KernelChaos':
      icon = <LinuxKernelIcon />
      break
    case 'NetworkChaos':
      icon = <NetworkIcon />
      break
    case 'PodChaos':
      icon = <PodLifecycleIcon />
      break
    case 'ProcessChaos':
      icon = <ProcessIcon />
      break
    case 'StressChaos':
      icon = <StressIcon />
      break
    case 'TimeChaos':
    case 'Schedule':
      icon = <ClockIcon />
      break
    case 'Serial':
      return <LinearScaleIcon fontSize={size} />
    case 'Parallel':
      return <LinearScaleIcon fontSize={size} sx={{ transform: 'rotate(90deg)', transformOrigin: 'center' }} />
    case 'Suspend':
      return <TimelapseIcon fontSize={size} />
  }

  return <SvgIcon fontSize={size}>{icon}</SvgIcon>
}

export function transByKind(kind: ExperimentKind | 'Workflow' | 'Schedule') {
  let id: string

  if (kind === 'Workflow') {
    id = 'workflows.title'
  } else if (kind === 'Schedule') {
    id = 'schedules.title'
  } else {
    id = `newE.target.${kind.replace('Chaos', '').toLowerCase()}.title`
  }

  return i18n(id)
}
