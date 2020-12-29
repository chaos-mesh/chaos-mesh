export function toTitleCase(s: string) {
  return s.charAt(0).toUpperCase() + s.substr(1)
}

export function joinObjKVs(obj: Record<string, string[]>, separator: string, filters?: string[]) {
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

export function difference<T>(setA: Set<T>, setB: Set<T>) {
  if (setA.size < setB.size) {
    ;[setA, setB] = [setB, setA]
  }
  const _difference = setA
  for (let el of setB) {
    _difference.delete(el)
  }
  return _difference
}

export function assumeType<T>(x: unknown): asserts x is T {
  return
}
