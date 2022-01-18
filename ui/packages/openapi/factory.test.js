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

import { typeTextToFieldType, typeTextToInitialValue } from './factory.js'

describe('openapi => factory', () => {
  test('typeTextToFieldType', () => {
    expect(typeTextToFieldType('string')).toBe('text')
    expect(typeTextToFieldType('Array<string>')).toBe('label')
    expect(typeTextToFieldType('{ [key: string]: string }')).toBe('string-string')
    expect(typeTextToFieldType('{ [key: string]: Array<string> }')).toBe('string-label')
    expect(typeTextToFieldType('number')).toBe('number')
    expect(typeTextToFieldType('boolean')).toBe('select')
    expect(() => typeTextToFieldType('unknown')).toThrowError()
  })

  test('typeTextToInitialValue', () => {
    expect(typeTextToInitialValue('string').text).toBe('')
    expect(typeTextToInitialValue('Array<string>').elements).toBeInstanceOf(Array)
    expect(typeTextToInitialValue('{ [key: string]: string }').properties.length).toBe(0)
    expect(typeTextToInitialValue('number').text).toBe('0')
    expect(typeTextToInitialValue('boolean').elements).toBeInstanceOf(Array)
    expect(() => typeTextToInitialValue('other')).toThrowError()
  })
})
