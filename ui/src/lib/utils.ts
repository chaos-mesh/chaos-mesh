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
export function toTitleCase(s: string) {
  return s.charAt(0).toUpperCase() + s.substr(1)
}

export function truncate(s: string) {
  if (s.length > 25) {
    return s.substring(0, 25) + '...'
  }

  return s
}

export function objToArrBySep(obj: Record<string, string | string[]>, separator: string, filters?: string[]) {
  return Object.entries(obj)
    .filter((d) => !filters?.includes(d[0]))
    .reduce(
      (acc: string[], [key, val]) =>
        acc.concat(Array.isArray(val) ? val.map((d) => `${key}${separator}${d}`) : `${key}${separator}${val}`),
      []
    )
}

export function arrToObjBySep(arr: string[], sep: string) {
  const result: any = {}

  arr.forEach((d) => {
    const split = d.split(sep)

    result[split[0]] = split[1]
  })

  return result as object
}
