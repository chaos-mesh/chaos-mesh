import { Event } from 'api/events.type'
import day from './dayjs'

const searchRegex = /(namespace:\S+)?\s?(kind:\S+)?\s?(time:\d{1,4}-\d{1,2}-\d{1,2})?\s?(.*)/

function parseSearch(search: string) {
  const matches = search.match(searchRegex)!
  console.log(matches)
  const namespace = matches[1] ? matches[1].split(':')[1].toLowerCase() : undefined
  const kind = matches[2] ? matches[2].split(':')[1].toLowerCase() : undefined
  const time = matches[3] ? matches[3].split(':')[1].toLowerCase() : undefined
  const content = matches[4].toLowerCase()

  return {
    namespace,
    kind,
    time,
    content,
  }
}

export function searchEvents(events: Event[], search: string) {
  let result = events
  const { namespace, kind, time, content } = parseSearch(search)

  if (namespace) {
    result = result.filter((d) => d.Namespace.toLowerCase().includes(namespace))
  }

  if (kind) {
    result = result.filter((d) => d.Kind.toLowerCase().includes(kind))
  }

  if (time) {
    const dayTime = day(time)

    if (dayTime.isValid()) {
      result = result.filter((d) => day(d.StartTime).isBefore(dayTime))
    }
  }

  if (content) {
    result = result.filter((d) => d.Experiment.toLowerCase().includes(content))
  }

  return result
}
