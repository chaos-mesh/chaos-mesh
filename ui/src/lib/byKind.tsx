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

import { ReactComponent as AWSIcon } from 'images/chaos/aws.svg'
import { ReactComponent as ClockIcon } from 'images/chaos/time.svg'
import { ReactComponent as DNSIcon } from 'images/chaos/dns.svg'
import { ReactComponent as DiskIcon } from 'images/chaos/disk.svg'
import { ExperimentKind } from 'components/NewExperiment/types'
import { ReactComponent as FileSystemIOIcon } from 'images/chaos/io.svg'
import { ReactComponent as GCPIcon } from 'images/chaos/gcp.svg'
import { ReactComponent as JavaIcon } from 'images/chaos/java.svg'
import { ReactComponent as K8SIcon } from 'images/k8s.svg'
import { ReactComponent as LinuxKernelIcon } from 'images/chaos/kernel.svg'
import { ReactComponent as NetworkIcon } from 'images/chaos/network.svg'
import { ReactComponent as PhysicIcon } from 'images/physic.svg'
import { ReactComponent as PodLifecycleIcon } from 'images/chaos/pod.svg'
import { ReactComponent as ProcessIcon } from 'images/chaos/process.svg'
import { ReactComponent as StressIcon } from 'images/chaos/stress.svg'
import { SvgIcon } from '@material-ui/core'
import T from 'components/T'

export function iconByKind(
  kind: ExperimentKind | 'Schedule' | 'k8s' | 'physic' | 'DiskChaos' | 'ProcessChaos',
  size: 'small' | 'large' = 'large'
) {
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
    case 'DiskChaos':
      icon = <DiskIcon />
      break
    case 'DNSChaos':
      icon = <DNSIcon />
      break
    case 'GCPChaos':
      icon = <GCPIcon />
      break
    case 'IOChaos':
      icon = <FileSystemIOIcon />
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

  return T(id)
}
