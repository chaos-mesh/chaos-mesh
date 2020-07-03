import { Experiment, ExperimentTarget, FormikCtx } from 'components/NewExperiment/types'

import { defaultExperimentSchema } from 'components/NewExperiment/constants'

export const ChaosKindKeyMap: {
  [kind: string]: { [key: string]: Exclude<keyof ExperimentTarget, 'kind'> }
} = {
  PodChaos: { key: 'pod_chaos' },
  NetworkChaos: { key: 'network_chaos' },
  IoChaos: { key: 'io_chaos' },
  KernelChaos: { key: 'kernel_chaos' },
  TimeChaos: { key: 'time_chaos' },
  StressChaos: { key: 'stress_chaos' },
}

export function parseSubmitValues(e: Experiment) {
  const values = JSON.parse(JSON.stringify(e))

  // Parse phase_selectors
  const phaseSelectors = values.scope.phase_selectors
  if (phaseSelectors.length === 1 && phaseSelectors[0] === 'all') {
    values.scope.phase_selectors = []
  }

  // Parse labels, label_selectors, annotations and annotation_selectors to object
  function helper1(selectors: string[]) {
    return selectors.reduce((acc: { [key: string]: string }, d) => {
      const splited = d.split(': ')

      acc[splited[0]] = splited[1]

      return acc
    }, {})
  }
  values.labels = helper1(values.labels)
  values.annotations = helper1(values.annotations)
  values.scope.label_selectors = helper1(values.scope.label_selectors as string[])
  values.scope.annotation_selectors = helper1(values.scope.annotation_selectors as string[])

  // Remove unrelated chaos
  const kind = values.target.kind
  Object.entries(ChaosKindKeyMap)
    .filter((k) => k[0] !== kind)
    .map((k) => k[1].key)
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

export function resetOtherChaos(formProps: FormikCtx, kind: string, action: string) {
  const { values, setFieldValue } = formProps

  const selectedChaosKind = kind
  const selectedChaosKey = ChaosKindKeyMap[selectedChaosKind].key

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

export function yamlToExperiments(yamlObj: any): Experiment {
  const { kind, metadata, spec } = yamlObj

  let halfResult = {
    ...defaultExperimentSchema,
    ...metadata,
    scope: {
      ...defaultExperimentSchema.scope,
      label_selectors: spec.selector.labelSelectors
        ? Object.entries(spec.selector.labelSelectors).map(([key, val]) => `${key}: ${val}`)
        : [],
      annotation_selectors: spec.selector.annotationSelectors
        ? Object.entries(spec.selector.annotationSelectors).map(([key, val]) => `${key}: ${val}`)
        : [],
      mode: spec.mode,
    },
    scheduler: {
      cron: spec.scheduler.cron,
      duration: spec.duration,
    },
  }

  delete spec.selector
  delete spec.mode
  delete spec.scheduler
  delete spec.duration

  const action = Object.keys(spec).filter((k) => k === spec.action)[0]

  return {
    ...halfResult,
    target: {
      ...defaultExperimentSchema.target,
      kind,
      [ChaosKindKeyMap[kind].key]: {
        action: spec.action,
        [action]: spec[action],
      },
    },
  }
}
