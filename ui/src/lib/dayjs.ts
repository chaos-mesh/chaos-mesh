import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(relativeTime)

export function dayComparator(a: string, b: string) {
  const dayA = dayjs(a)
  const dayB = dayjs(b)

  if (dayB.isAfter(dayA)) {
    return 1
  }

  if (dayB.isBefore(dayA)) {
    return -1
  }

  return 0
}

export default dayjs
