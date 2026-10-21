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
import ts from 'typescript'

import { typeTextToFieldType, typeTextToInitialValue } from './factory.js'

describe('openapi => factory', () => {
  test('typeTextToFieldType', () => {
    assert.equal(typeTextToFieldType('string'), 'text')
    assert.equal(typeTextToFieldType('string[]'), 'label')
    assert.equal(typeTextToFieldType('number[]'), 'numbers')
    assert.equal(typeTextToFieldType('{ [key: string]: string }'), 'text-text')
    assert.equal(typeTextToFieldType('string[][]'), 'text-label')
    assert.equal(typeTextToFieldType('{ [key: string]: string[] }'), 'text-label')
    assert.equal(typeTextToFieldType('number'), 'number')
    assert.equal(typeTextToFieldType('boolean'), 'select')
    assert.throws(() => typeTextToFieldType('unknown'))
  })

  test('typeTextToInitialValue', () => {
    assert.equal(typeTextToInitialValue('string').text, '')
    assert.ok(typeTextToInitialValue('string[]').elements instanceof Array)
    assert.ok(typeTextToInitialValue('number[]').elements instanceof Array)
    assert.equal(typeTextToInitialValue('string[][]').properties.length, 0)
    assert.equal(typeTextToInitialValue('{ [key: string]: string }').properties.length, 0)
    assert.equal(typeTextToInitialValue('{ [key: string]: string[] }').properties.length, 0)
    assert.equal(typeTextToInitialValue('number').text, '0')
    assert.equal(typeTextToInitialValue('boolean').kind, ts.SyntaxKind.FalseKeyword)
    assert.throws(() => typeTextToInitialValue('other'))
  })
})
