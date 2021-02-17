import { CallchainFrame, Experiment, ExperimentScope } from 'components/NewExperiment/types'

import _snakecase from 'lodash.snakecase'
import basic from 'components/NewExperimentNext/data/basic'
import snakeCaseKeys from 'snakecase-keys'

export function parseSubmit(e: Experiment) {
  const values: Experiment = JSON.parse(JSON.stringify(e))

  // Set default namespace when it's not present
  if (!values.namespace) {
    values.namespace = values.scope.namespace_selectors[0]
  }

  // Parse labels, label_selectors, annotations and annotation_selectors to object
  function helper1(selectors: string[], updateVal?: (s: string) => any) {
    return selectors.reduce((acc: Record<string, any>, d) => {
      const splited = d.replace(/\s/g, '').split(/:(.+)/)

      acc[splited[0]] = typeof updateVal === 'function' ? updateVal(splited[1]) : splited[1]

      return acc
    }, {})
  }
  // For parse scope
  function helper2(scope: ExperimentScope) {
    scope.label_selectors = helper1(scope.label_selectors as string[])
    scope.annotation_selectors = helper1(scope.annotation_selectors as string[])
    scope.pods = ((scope.pods as unknown) as string[]).reduce((acc, d) => {
      const [namespace, name] = d.split(':')
      if (acc.hasOwnProperty(namespace)) {
        acc[namespace].push(name)
      } else {
        acc[namespace] = [name]
      }

      return acc
    }, {} as Record<string, string[]>)

    // Parse phase_selectors
    const phaseSelectors = scope.phase_selectors
    if (phaseSelectors.length === 1 && phaseSelectors[0] === 'all') {
      scope.phase_selectors = []
    }
  }
  values.labels = helper1(values.labels as string[])
  values.annotations = helper1(values.annotations as string[])
  helper2(values.scope)

  const kind = values.target.kind

  // Handle NetworkChaos target
  if (kind === 'NetworkChaos') {
    const networkTarget = values.target.network_chaos.target_scope

    if (networkTarget) {
      if (networkTarget.mode) {
        helper2(values.target.network_chaos.target_scope!)
      } else {
        values.target.network_chaos.target_scope = undefined
      }
    }
  }

  if (kind === 'IoChaos' && values.target.io_chaos.action === 'attrOverride') {
    values.target.io_chaos.attr = helper1(values.target.io_chaos.attr as string[], (s: string) => parseInt(s, 10))
  }

  return values
}

function selectorsToArr(selectors: Object, separator: string) {
  return Object.entries(selectors).map(([key, val]) => `${key}${separator}${val}`)
}

export function yamlToExperiment(yamlObj: any): any {
  const { kind, metadata, spec } = snakeCaseKeys(yamlObj, {
    exclude: [/\.|\//], // Keys like app.kubernetes.io/component should be ignored
  }) as any

  if (!kind || !metadata || !spec) {
    throw new Error('Fail to parse the YAML file. Please check the kind, metadata, and spec fields.')
  }

  let result = {
    basic: {
      ...basic,
      ...metadata,
      labels: metadata.labels ? selectorsToArr(metadata.labels, ':') : [],
      annotations: metadata.annotations ? selectorsToArr(metadata.annotations, ':') : [],
      scope: {
        ...basic.scope,
        namespace_selectors: spec.selector.namespaces ?? [],
        label_selectors: spec.selector?.label_selectors ? selectorsToArr(spec.selector.label_selectors, ': ') : [],
        annotation_selectors: spec.selector?.annotation_selectors
          ? selectorsToArr(spec.selector.annotation_selectors, ': ')
          : [],
        mode: spec.mode ?? 'one',
        value: spec.value?.toString() ?? '',
      },
      scheduler: {
        cron: spec.scheduler?.cron ?? '',
        duration: spec.duration ?? '',
      },
    },
  }

  delete spec.selector
  delete spec.mode
  delete spec.value
  delete spec.scheduler
  delete spec.duration

  if (kind === 'NetworkChaos') {
    if (spec.target) {
      const namespace_selectors = spec.target.selector?.namespaces ?? []
      const label_selectors = spec.target.selector?.label_selectors
        ? selectorsToArr(spec.target.selector.label_selectors, ': ')
        : []
      const annotation_selectors = spec.target.selector?.annotation_selectors
        ? selectorsToArr(spec.target.selector.annotation_selectors, ': ')
        : []

      spec.target.selector && delete spec.target.selector

      spec.target_scope = {
        ...basic.scope,
        ...spec.target,
        namespace_selectors,
        label_selectors,
        annotation_selectors,
      }

      delete spec.target
    }
  }

  if (kind === 'IoChaos' && spec.attr) {
    spec.attr = selectorsToArr(spec.attr, ':')
  }

  if (kind === 'KernelChaos' && spec.fail_kern_request) {
    spec.fail_kern_request.callchain = spec.fail_kern_request.callchain.map((frame: CallchainFrame) => {
      if (!frame.parameters) {
        frame.parameters = ''
      }

      if (!frame.predicate) {
        frame.predicate = ''
      }

      return frame
    })
  }

  if (kind === 'StressChaos') {
    spec.stressors.cpu = {
      workers: 0,
      load: 0,
      options: [],
      ...spec.stressors.cpu,
    }

    spec.stressors.memory = {
      workers: 0,
      options: [],
      ...spec.stressors.memory,
    }
  }

  return {
    ...result,
    target: {
      kind,
      [_snakecase(kind)]: spec,
    },
  }
}
