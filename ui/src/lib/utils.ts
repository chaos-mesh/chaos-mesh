import { Experiment } from 'components/NewExperiment/types'

export function upperFirst(s: string) {
  if (!s) return ''

  return s.charAt(0).toUpperCase() + s.slice(1)
}

export const tabKinds = [
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
  const phaseSelectors = values.scope.phase_selectors
  if (phaseSelectors.length === 1 && phaseSelectors[0] === 'all') {
    values.scope.phase_selectors = []
  }

  const labelSelectors = values.scope.label_selectors
  values.scope.label_selectors = JSON.parse(labelSelectors)

  const kind = values.target.kind
  for (const key in values.target) {
    if (key === 'kind') {
      continue
    }

    if (
      tabKinds
        .filter((k) => k.kind !== kind)
        .map((k) => k.key)
        .includes(key)
    ) {
      delete (values.target as any)[key]
    }
  }

  for (const key in values.target) {
    if (key !== 'kind') {
      const chaos = (values.target as any)[key]

      for (const chaosKey in chaos) {
        if (chaosKey === 'action') {
          continue
        }

        if (chaosKey !== chaos.action) {
          delete chaos[chaosKey]
        }
      }
    }
  }

  return values
}
