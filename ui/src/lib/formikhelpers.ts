import { Experiment, ExperimentKind, Frame, Scope } from 'components/NewExperiment/types'
import { toCamelCase, toTitleCase } from './utils'

import { Template } from 'slices/workflows'
import { WorkflowBasic } from 'components/NewWorkflow'
import _snakecase from 'lodash.snakecase'
import basic from 'components/NewExperimentNext/data/basic'
import yaml from 'js-yaml'

export function parseSubmit<K extends ExperimentKind>(kind: K, e: Experiment<K>) {
  const values: Experiment<K> = JSON.parse(JSON.stringify(e))
  let { metadata, spec } = values

  // Set default namespace when it's not present
  if (!metadata.namespace) {
    metadata.namespace = spec.selector.namespaces[0]
  }

  // Parse labels, annotations, labelSelectors, and annotationSelectors to object
  function helper1(selectors: string[], updateVal?: (s: string) => any) {
    return selectors.reduce((acc: Record<string, any>, d) => {
      const splited = d.replace(/\s/g, '').split(/:(.+)/)

      acc[splited[0]] = typeof updateVal === 'function' ? updateVal(splited[1]) : splited[1]

      return acc
    }, {})
  }
  // Parse selector
  function helper2(scope: Scope['selector']) {
    if (scope.labelSelectors?.length) {
      scope.labelSelectors = helper1(scope.labelSelectors) as any
    } else {
      delete scope.labelSelectors
    }
    if (scope.annotationSelectors?.length) {
      scope.annotationSelectors = helper1(scope.annotationSelectors) as any
    } else {
      delete scope.annotationSelectors
    }

    // Parse phaseSelectors
    const phaseSelectors = scope.phaseSelectors
    if (phaseSelectors?.length === 1 && phaseSelectors[0] === 'all') {
      delete scope.phaseSelectors
    }

    // Parse pods
    if (scope.pods?.length) {
      scope.pods = scope.pods.reduce((acc, d) => {
        const [namespace, name] = d.split(':')
        if (acc.hasOwnProperty(namespace)) {
          acc[namespace].push(name)
        } else {
          acc[namespace] = [name]
        }

        return acc
      }, {} as Record<string, string[]>) as any
    } else {
      delete scope.pods
    }
  }
  if (metadata.labels?.length) {
    metadata.labels = helper1(metadata.labels) as any
  } else {
    delete metadata.labels
  }
  if (metadata.annotations?.length) {
    metadata.annotations = helper1(metadata.annotations) as any
  } else {
    delete metadata.annotations
  }
  helper2(spec.selector)

  if (kind === 'NetworkChaos') {
    if ((spec as any).target) {
      if (spec.mode) {
        helper2((spec as any).target)
      } else {
        ;(spec as any).target = undefined
      }
    }
  }

  if (kind === 'IOChaos' && (spec as any).action === 'attrOverride') {
    ;(spec as any).attr = helper1((spec as any).attr as string[], (s: string) => parseInt(s, 10))
  }

  return {
    apiVersion: 'chaos-mesh.org/v1alpha1',
    kind,
    metadata,
    spec,
  }
}

function selectorsToArr(selectors: Object, separator: string) {
  return Object.entries(selectors).map(([key, val]) => `${key}${separator}${val}`)
}

export function yamlToExperiment(yamlObj: any): any {
  const { kind, metadata, spec } = yamlObj

  if (!kind || !metadata || !spec) {
    throw new Error('Fail to parse the YAML file. Please check the kind, metadata, and spec fields.')
  }

  let result = {
    basic: {
      ...basic,
      ...metadata,
      labels: metadata.labels ? selectorsToArr(metadata.labels, ':') : [],
      annotations: metadata.annotations ? selectorsToArr(metadata.annotations, ':') : [],
      spec: {
        selector: {
          ...basic.spec.selector,
          namespaces: spec.selector.namespaces,
          labelSelectors: spec.selector?.labelSelectors ? selectorsToArr(spec.selector.labelSelectors, ': ') : [],
          annotation_selectors: spec.selector?.annotationSelectors
            ? selectorsToArr(spec.selector.annotationSelectors, ': ')
            : [],
        },
        mode: spec.mode ?? 'one',
        value: spec.value ?? '',
        duration: spec.duration ?? '',
      },
    },
  }

  delete spec.selector
  delete spec.mode
  delete spec.value
  delete spec.duration

  if (kind === 'NetworkChaos') {
    if (spec.target) {
      spec.target.selector.labelSelectors = spec.target.selector.labelSelectors
        ? selectorsToArr(spec.target.selector.labelSelectors, ': ')
        : []
      spec.target.selector.annotationSelectors = spec.target.selector.annotationSelectors
        ? selectorsToArr(spec.target.selector.annotationSelectors, ': ')
        : []
    }
  }

  if (kind === 'IOChaos' && spec.attr) {
    spec.attr = selectorsToArr(spec.attr, ':')
  }

  if (kind === 'KernelChaos' && spec.fail_kern_request) {
    spec.fail_kern_request.callchain = spec.fail_kern_request.callchain.map((frame: Frame) => {
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
      spec,
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

// FIXME
function scopeToYAMLJSON(scope: Scope['selector']) {
  return scope
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
