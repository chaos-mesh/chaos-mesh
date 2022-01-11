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
