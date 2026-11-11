/*
 * Copyright 2025 Chaos Mesh Authors.
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
const scalarFilters = new Set(['namespace', 'name', 'kind', 'object_id'])

export default function legacyEvents(spec) {
  const events = spec.paths?.['/events']?.get
  if (!events) {
    return spec
  }

  // Swagger 2 cannot describe the page-dependent response. The generated client retains the array contract.
  const parameters = events.parameters
    .filter((parameter) => parameter.in !== 'query' || !['page', 'include_total'].includes(parameter.name))
    .map((parameter) => {
      if (parameter.in !== 'query' || !scalarFilters.has(parameter.name)) {
        return parameter
      }

      const { items, collectionFormat, ...scalar } = parameter
      return { ...scalar, type: 'string' }
    })

  return {
    ...spec,
    paths: {
      ...spec.paths,
      '/events': {
        ...spec.paths['/events'],
        get: { ...events, parameters },
      },
    },
  }
}
