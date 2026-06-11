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

const order = ['apiVersion', 'kind', 'metadata', 'spec', 'name', 'namespace', 'labels', 'annotations']

export function dump(object: unknown, options?: jsyaml.DumpOptions): string {
  return jsyaml.dump(object, {
    sortKeys: (a, b) => {
      const indexA = order.indexOf(a)
      const indexB = order.indexOf(b)

      if (indexA !== -1 && indexB !== -1) {
        return indexA - indexB
      }
      if (indexA !== -1) return -1
      if (indexB !== -1) return 1

      return a.localeCompare(b)
    },
    ...options,
  })
}

export function load(str: string, options?: jsyaml.LoadOptions): unknown {
  return jsyaml.load(str, options)
}

const yaml = {
  dump,
  load,
}

export default yaml
