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
import legacyEvents from './legacy-events.js'

test('keeps the generated events array API compatible without changing other operations', () => {
  const filters = ['namespace', 'name', 'kind', 'object_id'].map((name) => ({
    in: 'query',
    name,
    type: 'array',
    items: { type: 'string' },
    collectionFormat: 'multi',
    description: name,
  }))
  const controls = [
    { in: 'query', name: 'page', type: 'integer' },
    { in: 'query', name: 'include_total', type: 'boolean' },
  ]
  const legacyParameters = [
    { in: 'query', name: 'limit', type: 'integer' },
    { in: 'query', name: 'start', type: 'string' },
    { in: 'query', name: 'end', type: 'string' },
    { in: 'query', name: 'sort', type: 'string' },
    { in: 'query', name: 'order', type: 'string' },
  ]
  const operation = {
    parameters: [...filters, ...controls, ...legacyParameters],
    responses: { 200: { schema: { type: 'array', items: { $ref: '#/definitions/core.Event' } } } },
  }
  const spec = { swagger: '2.0', paths: { '/events': { get: operation }, '/other': { get: operation } } }
  const result = legacyEvents(spec)

  expect(result.paths['/events'].get.parameters).toEqual([
    ...['namespace', 'name', 'kind', 'object_id'].map((name) => ({
      in: 'query',
      name,
      type: 'string',
      description: name,
    })),
    ...legacyParameters,
  ])
  expect(result.paths['/events'].get.responses).toEqual(operation.responses)
  expect(result.paths['/other'].get).toEqual(operation)
  expect(spec.paths['/events'].get.parameters).toEqual([...filters, ...controls, ...legacyParameters])
})
