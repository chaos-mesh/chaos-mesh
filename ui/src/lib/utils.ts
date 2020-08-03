export function toTitleCase(s: string) {
  return s.replace(/\w\S*/g, function (txt) {
    return txt.charAt(0).toUpperCase() + txt.substr(1).toLowerCase()
  })
}

export function joinObjKVs(obj: { [key: string]: string[] }, separator: string, filters?: string[]) {
  return Object.entries(obj)
    .filter((d) => !filters?.includes(d[0]))
    .reduce((acc: string[], [key, val]) => acc.concat(val.map((d) => `${key}${separator}${d}`)), [])
}
