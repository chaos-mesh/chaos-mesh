import { ReactComponent as AWSIcon } from 'images/chaos/aws.svg'
import { ReactComponent as ClockIcon } from 'images/chaos/time.svg'
import { ReactComponent as DNSIcon } from 'images/chaos/dns.svg'
import { ExperimentKind } from 'components/NewExperiment/types'
import { ReactComponent as FileSystemIOIcon } from 'images/chaos/io.svg'
import { ReactComponent as GCPIcon } from 'images/chaos/gcp.svg'
import { ReactComponent as LinuxKernelIcon } from 'images/chaos/kernel.svg'
import { ReactComponent as NetworkIcon } from 'images/chaos/network.svg'
import { ReactComponent as PodLifecycleIcon } from 'images/chaos/pod.svg'
import ScheduleIcon from '@material-ui/icons/Schedule'
import { ReactComponent as StressIcon } from 'images/chaos/stress.svg'
import { SvgIcon } from '@material-ui/core'
import T from 'components/T'

export function iconByKind(kind: ExperimentKind | 'Schedule', size: 'small' | 'large' = 'large') {
  let icon

  switch (kind) {
    case 'Schedule':
      icon = <ScheduleIcon />
      break
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
    case 'AwsChaos':
      icon = <AWSIcon />
      break
    case 'GcpChaos':
      icon = <GCPIcon />
      break
  }

  return kind !== 'Schedule' ? <SvgIcon fontSize={size}>{icon}</SvgIcon> : icon
}

export function transByKind(kind: ExperimentKind) {
  return T(`newE.target.${kind.replace('Chaos', '').toLowerCase()}.title`)
}
