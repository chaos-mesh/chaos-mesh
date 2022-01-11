import { typeTextToFieldType, typeTextToInitialValue } from './factory.js'

describe('openapi => factory', () => {
  test('typeTextToFieldType', () => {
    expect(typeTextToFieldType('string')).toBe('text')
    expect(typeTextToFieldType('Array<string>')).toBe('label')
    expect(typeTextToFieldType('string[]')).toBe('label')
    expect(typeTextToFieldType('unknown')).toBe('text')
  })

  test('typeTextToInitialValue', () => {
    expect(typeTextToInitialValue('string').text).toBe('')
    expect(typeTextToInitialValue('Array<string>').elements).toBeInstanceOf(Array)
    expect(typeTextToInitialValue('string[]').elements).toBeInstanceOf(Array)
    expect(typeTextToInitialValue('number').text).toBe('0')
    expect(typeTextToInitialValue('other').text).toBe('')
  })
})
