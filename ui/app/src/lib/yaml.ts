/*
 * Copyright 2026 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import jsyaml from 'js-yaml'

/**
 * Recursively orders the keys of a Kubernetes resource object.
 *
 * It ensures canonical ordering for:
 * 1. Resource root level (including any nested objects representing embedded
 *    Kubernetes resources that have both `apiVersion` and `kind` fields):
 *    apiVersion, kind, metadata, spec, status
 * 2. Metadata level: name, namespace, labels, annotations
 *
 * For all other fields and nested objects (such as those inside `spec` or `status`),
 * it preserves the original key insertion order. This avoids unwanted global/recursive
 * sorting side-effects that can make PR diffs noisy.
 */
function sortKeysForKubernetes(val: any, path: string[] = []): any {
  if (val === null || typeof val !== 'object') {
    return val
  }

  if (Array.isArray(val)) {
    return val.map((item) => sortKeysForKubernetes(item, path))
  }

  let currentPath = path
  if ('apiVersion' in val && 'kind' in val) {
    currentPath = []
  }

  const keys = Object.keys(val)
  const sortedKeys = [...keys]

  if (currentPath.length === 0) {
    const rootOrder = ['apiVersion', 'kind', 'metadata', 'spec', 'status']
    sortedKeys.sort((a, b) => {
      const idxA = rootOrder.indexOf(a)
      const idxB = rootOrder.indexOf(b)
      if (idxA !== -1 && idxB !== -1) return idxA - idxB
      if (idxA !== -1) return -1
      if (idxB !== -1) return 1
      return 0
    })
  } else if (currentPath.length === 1 && currentPath[0] === 'metadata') {
    const metadataOrder = ['name', 'namespace', 'labels', 'annotations']
    sortedKeys.sort((a, b) => {
      const idxA = metadataOrder.indexOf(a)
      const idxB = metadataOrder.indexOf(b)
      if (idxA !== -1 && idxB !== -1) return idxA - idxB
      if (idxA !== -1) return -1
      if (idxB !== -1) return 1
      return 0
    })
  }

  const result: any = {}
  for (const key of sortedKeys) {
    result[key] = sortKeysForKubernetes(val[key], [...currentPath, key])
  }
  return result
}

export function dump(object: unknown, options?: jsyaml.DumpOptions): string {
  const sortedObject = sortKeysForKubernetes(object)
  // Omit sortKeys from options so that js-yaml uses insertion order of sortedObject
  const { sortKeys, ...restOptions } = options || {}
  return jsyaml.dump(sortedObject, restOptions)
}

export function load(str: string, options?: jsyaml.LoadOptions): unknown {
  return jsyaml.load(str, options)
}

const yaml = {
  dump,
  load,
}

export default yaml
