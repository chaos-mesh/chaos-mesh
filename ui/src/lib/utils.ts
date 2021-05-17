export function toTitleCase(s: string) {
  return s.charAt(0).toUpperCase() + s.substr(1)
}

export function truncate(s: string) {
  if (s.length > 7) {
    return s.substring(0, 15) + '...'
  }

  return s
}

export function joinObjKVs(obj: Record<string, string[]>, separator: string, filters?: string[]) {
  return Object.entries(obj)
    .filter((d) => !filters?.includes(d[0]))
    .reduce((acc: string[], [key, val]) => acc.concat(val.map((d) => `${key}${separator}${d}`)), [])
}

export function arrToObjBySep(arr: string[], sep: string) {
  const result: any = {}

  arr.forEach((d) => {
    const split = d.split(sep)

    result[split[0]] = split[1]
  })

  return result as object
}
