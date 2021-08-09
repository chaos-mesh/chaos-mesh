export function toTitleCase(s: string) {
  return s.charAt(0).toUpperCase() + s.substr(1)
}

export function toCamelCase(s: string) {
  return s.charAt(0).toLowerCase() + s.substr(1)
}

export function truncate(s: string) {
  if (s.length > 25) {
    return s.substring(0, 25) + '...'
  }

  return s
}

export function objToArrBySep(obj: Record<string, string | string[]>, separator: string, filters?: string[]) {
  return Object.entries(obj)
    .filter((d) => !filters?.includes(d[0]))
    .reduce(
      (acc: string[], [key, val]) =>
        acc.concat(Array.isArray(val) ? val.map((d) => `${key}${separator}${d}`) : `${key}${separator}${val}`),
      []
    )
}

export function arrToObjBySep(arr: string[], sep: string) {
  const result: any = {}

  arr.forEach((d) => {
    const split = d.split(sep)

    result[split[0]] = split[1]
  })

  return result as object
}
