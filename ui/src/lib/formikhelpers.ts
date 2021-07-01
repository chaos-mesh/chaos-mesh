import { CallchainFrame, Experiment, ExperimentScope } from 'components/NewExperiment/types'
import { arrToObjBySep, toCamelCase, toTitleCase } from './utils'

import { Template } from 'slices/workflows'
import { WorkflowBasic } from 'components/NewWorkflow'
import _snakecase from 'lodash.snakecase'
import basic from 'components/NewExperimentNext/data/basic'
import snakeCaseKeys from 'snakecase-keys'
import yaml from 'js-yaml'

export function parseSubmit(e: Experiment) {
  const values: Experiment = JSON.parse(JSON.stringify(e))

  // Set default namespace when it's not present
  if (!values.namespace) {
    values.namespace = values.scope.namespaces[0]
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
    scope.pods = (scope.pods as unknown as string[]).reduce((acc, d) => {
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

  if (kind === 'IOChaos' && values.target.io_chaos.action === 'attrOverride') {
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
        namespaces: spec.selector.namespaces ?? [],
        label_selectors: spec.selector?.label_selectors ? selectorsToArr(spec.selector.label_selectors, ': ') : [],
        annotation_selectors: spec.selector?.annotation_selectors
          ? selectorsToArr(spec.selector.annotation_selectors, ': ')
          : [],
        mode: spec.mode ?? 'one',
        value: spec.value?.toString() ?? '',
      },
      scheduler: {
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
      const namespaces = spec.target.selector?.namespaces ?? []
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
        namespaces,
        label_selectors,
        annotation_selectors,
      }

      delete spec.target
    }
  }

  if (kind === 'IOChaos' && spec.attr) {
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

function validate(defaultI18n: string, i18n?: string) {
  return function (value: string) {
    let error

    if (value === '') {
      error = i18n ?? defaultI18n
    }

    return error
  }
}
export const validateName = (i18n?: string) => validate('The name is required', i18n)
export const validateDuration = (i18n?: string) => validate('The duration is required', i18n)
export const validateDeadline = (i18n?: string) => validate('The deadline is required', i18n)
export const validateImage = (i18n?: string) => validate('The image is required', i18n)

function scopeToYAMLJSON(scope: ExperimentScope) {
  const result = {
    selector: {} as any,
    mode: scope.mode,
  }

  if (scope.namespaces.length) {
    result.selector.namespaces = scope.namespaces
  }

  if ((scope.label_selectors as string[]).length) {
    result.selector.labelSelectors = arrToObjBySep(scope.label_selectors as string[], ': ')
  }

  if ((scope.annotation_selectors as string[]).length) {
    result.selector.annotationSelectors = arrToObjBySep(scope.annotation_selectors as string[], ': ')
  }

  return result
}

export function constructWorkflow(basic: WorkflowBasic, templates: Template[]) {
  const { name, namespace, deadline } = basic
  const children: string[] = templates.map((d) => d.name)
  const realTemplates: Record<string, any>[] = []

  function pushTemplate(template: any) {
    if (!realTemplates.some((t) => t.name === template.name)) {
      realTemplates.push(template)
    }
  }

  function recurInsertTemplates(templates: Template[]) {
    templates.forEach((t) => {
      switch (t.type) {
        case 'single':
          const experiment = t.experiment!
          const basic = experiment.basic
          const kind = experiment.target.kind
          const spec = _snakecase(kind)

          pushTemplate({
            name: t.name,
            templateType: kind,
            deadline: experiment.basic.deadline,
            [toCamelCase(kind)]: {
              ...scopeToYAMLJSON(basic.scope),
              ...experiment.target[spec],
            },
          })

          break
        case 'serial':
        case 'parallel':
        case 'custom':
          t.children!.forEach((d) => {
            if (d.children) {
              pushTemplate({
                name: d.name,
                templateType: toTitleCase(d.type),
                deadline: d.deadline,
                children: d.children!.map((dd) => dd.name),
              })

              recurInsertTemplates(d.children)
            } else {
              if (d.type === 'suspend') {
                pushTemplate({
                  name: d.name,
                  templateType: 'Suspend',
                  deadline: d.deadline,
                })

                return
              }

              const e = d.experiment!
              const basic = e.basic
              const name = basic.name
              const kind = e.target.kind
              const spec = _snakecase(kind)

              pushTemplate({
                name,
                templateType: kind,
                deadline: e.basic.deadline,
                [toCamelCase(kind)]: {
                  ...scopeToYAMLJSON(basic.scope),
                  ...e.target[spec],
                },
              })
            }
          })

          pushTemplate({
            name: t.name,
            templateType: toTitleCase(t.type === 'custom' ? 'task' : t.type),
            deadline: t.type !== 'custom' ? t.deadline : undefined,
            children: t.type !== 'custom' ? t.children!.map((d) => d.name) : undefined,
            task:
              t.type === 'custom'
                ? {
                    container: t.custom?.container,
                  }
                : undefined,
            conditionalBranches: t.type === 'custom' ? t.custom?.conditionalBranches : undefined,
          })

          break
        case 'suspend':
          pushTemplate({
            name: t.name,
            templateType: 'Suspend',
            deadline: t.deadline,
          })

          break
        default:
          break
      }
    })
  }

  recurInsertTemplates(templates)

  return yaml.dump(
    {
      apiVersion: 'chaos-mesh.org/v1alpha1',
      kind: 'Workflow',
      metadata: {
        name,
        namespace,
      },
      spec: {
        entry: 'entry',
        templates: [
          {
            name: 'entry',
            templateType: 'Serial',
            deadline,
            children,
          },
          ...realTemplates,
        ],
      },
    },
    {
      replacer: (_, value) => {
        if (Array.isArray(value)) {
          return value.length ? value : undefined
        }

        switch (typeof value) {
          case 'string':
            return value !== '' ? value : undefined
          default:
            return value
        }
      },
    }
  )
}
