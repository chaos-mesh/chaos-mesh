/*
 * Copyright 2022 Chaos Mesh Authors.
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

import { cleanMarkers, getUIFormAction, getUIFormEnum, isUIFormIgnore } from './utils'

describe('openapi => utils', () => {
  describe('getUIFormEnum', () => {
    test('returns an array', () => {
      expect(
        getUIFormEnum(`
        /**
         * just a comment
         *
         * ui:form:enum=a;b;c
         */
      `)
      ).toEqual(['a', 'b', 'c'])
    })

    test('returns an empty array', () => {
      expect(
        getUIFormEnum(`
        /**
         *
         */
      `)
      ).toHaveLength(0)
    })
  })

  describe('getUIFormAction', () => {
    test('returns a action', () => {
      expect(
        getUIFormAction(`
        /**
         * ui:form:action=a
         */
      `)
      ).toBe('a')
    })

    test('returns an empty string', () => {
      expect(
        getUIFormAction(`
        /**
         *
         */
      `)
      ).toBe('')
    })
  })

  describe('isUIFormIgnore', () => {
    test('yes', () => {
      expect(
        isUIFormIgnore(`
        /**
         * ui:form:ignore
         */
      `)
      ).toBe(true)
    })

    test('no', () => {
      expect(
        isUIFormIgnore(`
        /**
         * ui:form:ig
         */
      `)
      ).toBe(false)
    })
  })

  test('cleanMarkers', () => {
    expect(cleanMarkers('DeviceName indicates the name of the device. ui:form:action=detach-volume +optional')).toBe(
      'Optional. DeviceName indicates the name of the device.'
    )
  })
})
