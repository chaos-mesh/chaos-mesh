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
