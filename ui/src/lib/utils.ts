export function upperFirst(s: string) {
  if (!s) return ''

  return s.charAt(0).toUpperCase() + s.slice(1)
}

export function joinObjKVs(obj: { [key: string]: string[] }, separator: string) {
  return Object.entries(obj).reduce(
    (acc: string[], [key, val]) => acc.concat(val.map((d) => `${key}${separator}${d}`)),
    []
  )
}
