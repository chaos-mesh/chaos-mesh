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

export function arrToObjBySep(arr: string[], sep: string) {
  const result: any = {}

  arr.forEach((d) => {
    const splited = d.split(sep)

    result[splited[0]] = splited[1]
  })

  return result as object
}
