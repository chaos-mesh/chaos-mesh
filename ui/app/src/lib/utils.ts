/*
 * Copyright 2021 Chaos Mesh Authors.
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
import _ from 'lodash'

export function objToArrBySep(obj: Record<string, string | string[]>, separator: string) {
  return Object.entries(obj).reduce<string[]>(
    (acc, [k, v]) => acc.concat(Array.isArray(v) ? v.map((d) => `${k}${separator}${d}`) : `${k}${separator}${v}`),
    []
  )
}

export function arrToObjBySep(
  arr: string[],
  sep: string,
  options?: { removeAllSpaces?: boolean; updateVal?: (s: string) => any }
) {
  return arr.reduce<Record<string, string>>((acc, d) => {
    let processed = d

    if (options?.removeAllSpaces) {
      processed = processed.replace(/\s/g, '')
    }

    let [k, v] = processed.split(sep)

    if (options?.updateVal) {
      v = options.updateVal(v)
    }

    acc[k] = v

    return acc
  }, {})
}

/**
 * Recursively check if a value is empty.
 *
 * @export
 * @param {*} value
 * @return {boolean}
 */
export function isDeepEmpty(value: any): boolean {
  if (!value) {
    return true
  }

  if (_.isArray(value) && _.isEmpty(value)) {
    return true
  }

  if (_.isObject(value)) {
    return _.every(value, isDeepEmpty)
  }

  return false
}

/**
 * Remove empty values from nested object.
 *
 * @export
 * @param {*} obj
 */
export function sanitize(obj: any) {
  return JSON.parse(JSON.stringify(obj, (_, value: any) => (isDeepEmpty(value) ? undefined : value)) ?? '{}')
}

export function concatKindAction(kind: string, action?: string) {
  return `${kind}${action ? ` / ${action}` : ''}`
}
