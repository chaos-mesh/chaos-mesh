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
import ts from 'typescript'

import { typeTextToFieldType, typeTextToInitialValue } from './factory.js'

describe('openapi => factory', () => {
  test('typeTextToFieldType', () => {
    expect(typeTextToFieldType('string')).toBe('text')
    expect(typeTextToFieldType('string[]')).toBe('label')
    expect(typeTextToFieldType('number[]')).toBe('numbers')
    expect(typeTextToFieldType('{ [key: string]: string }')).toBe('text-text')
    expect(typeTextToFieldType('string[][]')).toBe('text-label')
    expect(typeTextToFieldType('{ [key: string]: string[] }')).toBe('text-label')
    expect(typeTextToFieldType('number')).toBe('number')
    expect(typeTextToFieldType('boolean')).toBe('select')
    expect(() => typeTextToFieldType('unknown')).toThrowError()
  })

  test('typeTextToInitialValue', () => {
    expect(typeTextToInitialValue('string').text).toBe('')
    expect(typeTextToInitialValue('string[]').elements).toBeInstanceOf(Array)
    expect(typeTextToInitialValue('number[]').elements).toBeInstanceOf(Array)
    expect(typeTextToInitialValue('string[][]').properties.length).toBe(0)
    expect(typeTextToInitialValue('{ [key: string]: string }').properties.length).toBe(0)
    expect(typeTextToInitialValue('{ [key: string]: string[] }').properties.length).toBe(0)
    expect(typeTextToInitialValue('number').text).toBe('0')
    expect(typeTextToInitialValue('boolean').kind).toBe(ts.SyntaxKind.FalseKeyword)
    expect(() => typeTextToInitialValue('other')).toThrowError()
  })
})
