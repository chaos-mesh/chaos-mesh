import { Experiment, ExperimentTarget, StepperFormProps } from 'components/NewExperiment/types'

import { defaultExperimentSchema } from 'components/NewExperiment/constants'

const ChaosKindsAndKeys: {
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

  // Parse label_selectors and annotation_selectors to object
  function helper1(selectors: string[]) {
    return selectors.length > 0
      ? selectors.reduce((acc: { [key: string]: string }, d) => {
          const splited = d.split(': ')

          acc[splited[0]] = splited[1]

          return acc
        }, {})
      : {}
  }
  values.scope.label_selectors = helper1(values.scope.label_selectors as string[])
  values.scope.annotation_selectors = helper1(values.scope.annotation_selectors as string[])

  // Remove unrelated chaos
  const kind = values.target.kind
  ChaosKindsAndKeys.filter((k) => k.kind !== kind)
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

        // Handle container-kill action
        if (chaos.action === 'container-kill' && action === 'container_name') {
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

export function mustSchedule(formikValues: Experiment) {
  if (
    formikValues.target.pod_chaos.action === 'pod-kill' ||
    formikValues.target.pod_chaos.action === 'container-kill'
  ) {
    return true
  }

  return false
}

export function resetOtherChaos(formProps: StepperFormProps, kind: string, action: string) {
  const { values, setFieldValue } = formProps

  const selectedChaosKind = kind
  const selectedChaosKey = ChaosKindsAndKeys.filter((d) => d.kind === selectedChaosKind)[0].key

  const updatedTarget = {
    ...defaultExperimentSchema.target,
    ...{
      kind: selectedChaosKind,
      [selectedChaosKey]: {
        ...values.target[selectedChaosKey],
        ...{
          action,
        },
      },
    },
  }

  setFieldValue('target', updatedTarget)
}
