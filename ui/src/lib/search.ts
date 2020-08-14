import { Event } from 'api/events.type'

const searchRegex = /(namespace:\S+)?\s?(kind:\S+)?\s?(pod:\S+)?\s?(uuid:\S+)?\s?(.*)/

function parseSearch(search: string) {
  const matches = search.match(searchRegex)!
  const namespace = matches[1] ? matches[1].split(':')[1].toLowerCase() : undefined
  const kind = matches[2] ? matches[2].split(':')[1].toLowerCase() : undefined
  const pod = matches[3] ? matches[3].split(':')[1].toLowerCase() : undefined
  const uuid = matches[4] ? matches[4].split(':')[1].toLowerCase() : undefined
  const content = matches[5].toLowerCase()

  return {
    namespace,
    kind,
    pod,
    uuid,
    content,
  }
}

export function searchEvents(events: Event[], search: string) {
  let result = events
  const { namespace, kind, pod, uuid, content } = parseSearch(search)

  if (namespace) {
    result = result.filter((d) => d.namespace.toLowerCase().includes(namespace))
  }

  if (kind) {
    result = result.filter((d) => d.kind.toLowerCase().includes(kind))
  }

  if (pod) {
    result = result.filter((d) => d.pods?.some((d) => d.pod_name.toLowerCase().includes(pod)))
  }

  if (uuid) {
    result = result.filter((d) => d.experiment_id.toLowerCase().startsWith(uuid))
  }

  if (content) {
    result = result.filter((d) => d.experiment.toLowerCase().includes(content))
  }

  return result
}
