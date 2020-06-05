import { Experiment, ExperimentTarget } from 'components/NewExperiment/types'

export const targetVerticalTabsKinds: {
  kind: string
  key: Exclude<keyof ExperimentTarget, 'kind'>
}[] = [
  {
    kind: 'PodChaos',
    key: 'pod_chaos',
  },
  {
    kind: 'NetworkChaos',
    key: 'network_chaos',
  },
  {
    kind: 'IoChaos',
    key: 'io_chaos',
  },
  {
    kind: 'KernelChaos',
    key: 'kernel_chaos',
  },
  {
    kind: 'TimeChaos',
    key: 'time_chaos',
  },
  {
    kind: 'StressChaos',
    key: 'stress_chaos',
  },
]

export function parseSubmitValues(values: Experiment) {
  // Parse phase_selectors
  const phaseSelectors = values.scope.phase_selectors
  if (phaseSelectors.length === 1 && phaseSelectors[0] === 'all') {
    values.scope.phase_selectors = []
  }

  // Parse label_selectors to object
  let labelSelectors = values.scope.label_selectors
  try {
    labelSelectors = JSON.parse(labelSelectors as string)
  } catch {
    labelSelectors = {}
  }
  values.scope.label_selectors = labelSelectors

  // Remove unrelated chaos
  const kind = values.target.kind
  targetVerticalTabsKinds
    .filter((k) => k.kind !== kind)
    .map((k) => k.key)
    .forEach((k) => delete values.target[k])

  // Remove unrelated actions
  for (const key in values.target) {
    if (key !== 'kind') {
      const chaos = (values.target as any)[key]

      for (const action in chaos) {
        if (action === 'action') {
          continue
        }

        if (action !== chaos.action) {
          delete chaos[action]
        }
      }
    }
  }

  return values
}

export function mustImmediate(formikValues: Experiment) {
  if (
    formikValues.target.pod_chaos.action === 'pod-kill' ||
    formikValues.target.pod_chaos.action === 'container-kill'
  ) {
    return true
  }

  return false
}
