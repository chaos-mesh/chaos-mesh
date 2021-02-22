import { DateTime } from 'luxon'

export function comparator(a: string, b: string) {
  const da = DateTime.fromISO(a)
  const db = DateTime.fromISO(b)

  if (da > db) {
    return 1
  }

  if (da < db) {
    return -1
  }

  return 0
}

export const now = DateTime.local

export const format = (date: string, locale: string = 'en') =>
  DateTime.fromISO(date, { locale }).toFormat('yyyy-MM-dd HH:mm:ss a')

export default DateTime
