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

import { sanitize } from './utils'

test('sanitize an object', () => {
  expect(
    sanitize({
      a: 1,
      b: '',
      c: null,
      d: 'd',
    })
  ).toEqual({
    a: 1,
    d: 'd',
  })
})

test('sanitize an object where all values are empty', () => {
  expect(
    sanitize({
      a: 0,
      b: '',
      c: null,
      d: undefined,
      e: [],
    })
  ).toEqual({})
})
