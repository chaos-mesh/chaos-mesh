import { ReactComponent as AWSIcon } from 'images/chaos/aws.svg'
import { ReactComponent as ClockIcon } from 'images/chaos/time.svg'
import { ReactComponent as DNSIcon } from 'images/chaos/dns.svg'
import { ExperimentKind } from 'components/NewExperiment/types'
import { ReactComponent as FileSystemIOIcon } from 'images/chaos/io.svg'
import { ReactComponent as GCPIcon } from 'images/chaos/gcp.svg'
import { ReactComponent as LinuxKernelIcon } from 'images/chaos/kernel.svg'
import { ReactComponent as NetworkIcon } from 'images/chaos/network.svg'
import { ReactComponent as PodLifecycleIcon } from 'images/chaos/pod.svg'
import { ReactComponent as StressIcon } from 'images/chaos/stress.svg'
import { SvgIcon } from '@material-ui/core'
import T from 'components/T'

export function iconByKind(kind: ExperimentKind | 'Schedule', size: 'small' | 'large' = 'large') {
  let icon

  switch (kind) {
    case 'PodChaos':
      icon = <PodLifecycleIcon />
      break
    case 'NetworkChaos':
      icon = <NetworkIcon />
      break
    case 'IOChaos':
      icon = <FileSystemIOIcon />
      break
    case 'KernelChaos':
      icon = <LinuxKernelIcon />
      break
    case 'TimeChaos':
    case 'Schedule':
      icon = <ClockIcon />
      break
    case 'StressChaos':
      icon = <StressIcon />
      break
    case 'DNSChaos':
      icon = <DNSIcon />
      break
    case 'AWSChaos':
      icon = <AWSIcon />
      break
    case 'GCPChaos':
      icon = <GCPIcon />
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
