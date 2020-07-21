import { Event } from 'api/events.type'

const searchRegex = /(namespace:\S+)?\s?(kind:\S+)?\s?(.*)/

function parseSearch(search: string) {
  const matches = search.match(searchRegex)!
  const namespace = matches[1] ? matches[1].split(':')[1].toLowerCase() : undefined
  const kind = matches[2] ? matches[2].split(':')[1].toLowerCase() : undefined
  const content = matches[3].toLowerCase()

  return {
    namespace,
    kind,
    content,
  }
}

export function searchEvents(events: Event[], search: string) {
  let result = events
  const { namespace, kind, content } = parseSearch(search)

  if (namespace) {
    result = result.filter((d) => d.Namespace.toLowerCase().includes(namespace))
  }

  if (kind) {
    result = result.filter((d) => d.Kind.toLowerCase().includes(kind))
  }

  if (content) {
    result = result.filter((d) => d.Experiment.toLowerCase().includes(content))
  }

  return result
}
