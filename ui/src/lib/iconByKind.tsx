import { ReactComponent as ClockIcon } from 'images/chaos/time.svg'
import { ReactComponent as DNSIcon } from 'images/chaos/dns.svg'
import { ExperimentKind } from 'components/NewExperiment/types'
import { ReactComponent as FileSystemIOIcon } from 'images/chaos/io.svg'
import { ReactComponent as LinuxKernelIcon } from 'images/chaos/kernel.svg'
import { ReactComponent as NetworkIcon } from 'images/chaos/network.svg'
import { ReactComponent as PodLifecycleIcon } from 'images/chaos/pod.svg'
import React from 'react'
import { ReactComponent as StressIcon } from 'images/chaos/stress.svg'
import { SvgIcon } from '@material-ui/core'

export default function iconByKind(kind: ExperimentKind, size: 'small' | 'large' = 'large') {
  let icon

  switch (kind) {
    case 'PodChaos':
      icon = <PodLifecycleIcon />
      break
    case 'NetworkChaos':
      icon = <NetworkIcon />
      break
    case 'IoChaos':
      icon = <FileSystemIOIcon />
      break
    case 'KernelChaos':
      icon = <LinuxKernelIcon />
      break
    case 'TimeChaos':
      icon = <ClockIcon />
      break
    case 'StressChaos':
      icon = <StressIcon />
      break
    case 'DNSChaos':
      icon = <DNSIcon />
      break
    default:
      icon = <PodLifecycleIcon />
  }

  return <SvgIcon fontSize={size}>{icon}</SvgIcon>
}
