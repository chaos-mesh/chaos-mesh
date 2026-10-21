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
import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import { cleanMarkers, getUIFormEnum, getUIFormWhen, isUIFormIgnore } from './utils.js'

describe('openapi => utils', () => {
  describe('getUIFormEnum', () => {
    test('returns an array', () => {
      assert.deepEqual(
        getUIFormEnum(`
        /**
         * just a comment
         *
         * +ui:form:enum=a;b;c
         */
      `),
        ['a', 'b', 'c'],
      )

      assert.deepEqual(
        getUIFormEnum(`
        /**
         * just a comment
         *
         * +kubebuilder:validation:Enum=a;b;c
         */
      `),
        ['a', 'b', 'c'],
      )

      assert.deepEqual(
        getUIFormEnum(`
        /**
         * just a comment
         *
         * +kubebuilder:validation:Enum=a;b;c
         * +ui:form:enum=a;b;c
         */
      `),
        ['a', 'b', 'c'],
      )
    })

    test('returns an empty array', () => {
      assert.equal(
        getUIFormEnum(`
        /**
         *
         */
      `).length,
        0,
      )
    })
  })

  describe('getUIFormWhen', () => {
    test('returns an expression string', () => {
      assert.equal(
        getUIFormWhen(`
        /**
         * +ui:form:when=action=='a'
         */
      `),
        "action=='a'",
      )
    })

    test('returns false', () => {
      assert.equal(
        getUIFormWhen(`
        /**
         *
         */
      `),
        false,
      )
    })
  })

  describe('isUIFormIgnore', () => {
    test('yes', () => {
      assert.equal(
        isUIFormIgnore(`
        /**
         * +ui:form:ignore
         */
      `),
        true,
      )
    })

    test('no', () => {
      assert.equal(
        isUIFormIgnore(`
        /**
         * +ui:form:ig
         */
      `),
        false,
      )
    })
  })

  test('cleanMarkers', () => {
    assert.equal(
      cleanMarkers('DeviceName indicates the name of the device. +ui:form:when=action="detach-volume" +optional'),
      'Optional. DeviceName indicates the name of the device.',
    )
  })
})
