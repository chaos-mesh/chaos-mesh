import { ExperimentKind } from 'components/NewExperiment/types'

const mapping = new Map<ExperimentKind, string>([
  ['AWSChaos', 'awsChaos'],
  ['AzureChaos', 'azureChaos'],
  ['BlockChaos', 'blockChaos'],
  ['DNSChaos', 'dnsChaos'],
  ['GCPChaos', 'gcpChaos'],
  ['HTTPChaos', 'httpChaos'],
  ['IOChaos', 'ioChaos'],
  ['JVMChaos', 'jvmChaos'],
  ['K8SChaos', 'k8sChaos'],
  ['KernelChaos', 'kernelChaos'],
  ['NetworkChaos', 'networkChaos'],
  ['PhysicalMachineChaos', 'physicalmachineChaos'],
  ['PodChaos', 'podChaos'],
  ['StressChaos', 'stressChaos'],
  ['TimeChaos', 'timeChaos'],
])

export function templateTypeToFieldName(templateType: ExperimentKind): string {
  return mapping.get(templateType)!
}
