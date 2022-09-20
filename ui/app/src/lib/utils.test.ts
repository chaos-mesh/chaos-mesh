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
import { arrToObjBySep, isDeepEmpty, sanitize } from './utils'

describe('lib/utils', () => {
  describe('arrToObjBySep', () => {
    it('should convert array to object', () => {
      const arr = ['foo=bar', 'baz=qux']
      const result = arrToObjBySep(arr, '=')

      expect(result).toEqual({ foo: 'bar', baz: 'qux' })
    })

    it('should remove all spaces from the array items', () => {
      const arr = ['foo = bar', 'baz = qux ']
      const result = arrToObjBySep(arr, '=', { removeAllSpaces: true })

      expect(result).toEqual({ foo: 'bar', baz: 'qux' })
    })

    it('should convert array to object with value to a number', () => {
      const arr = ['foo=1', 'baz=2']
      const result = arrToObjBySep(arr, '=', { updateVal: (s) => +s })

      expect(result).toEqual({ foo: 1, baz: 2 })
    })
  })

  describe('isDeepEmpty', () => {
    it('checks some primitive values', () => {
      expect(isDeepEmpty(true)).toBeFalsy()
      expect(isDeepEmpty(false)).toBeTruthy()
      expect(isDeepEmpty(null)).toBeTruthy()
      expect(isDeepEmpty(undefined)).toBeTruthy()
      expect(isDeepEmpty(1)).toBeFalsy()
      expect(isDeepEmpty(0)).toBeTruthy()
      expect(isDeepEmpty(NaN)).toBeTruthy()
      expect(isDeepEmpty('string')).toBeFalsy()
      expect(isDeepEmpty('')).toBeTruthy()
    })

    it('checks arrays', () => {
      expect(isDeepEmpty([])).toBeTruthy()
      expect(isDeepEmpty([1])).toBeFalsy()
    })

    it('checks some objects', () => {
      expect(isDeepEmpty({})).toBeTruthy()
      expect(isDeepEmpty({ a: 1 })).toBeFalsy()
    })

    it('checks a nested object', () => {
      expect(isDeepEmpty({ a: { b: { c: {} } } })).toBeTruthy()
    })
  })

  describe('sanitize', () => {
    it('sanitizes an normal object', () => {
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

    it('sanitizes an object where all values are empty', () => {
      expect(
        sanitize({
          a: 0,
          b: '',
          c: null,
          d: undefined,
          e: [],
          f: {},
        })
      ).toEqual({})
    })
  })
})
