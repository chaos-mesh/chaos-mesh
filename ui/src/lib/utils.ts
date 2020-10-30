export function toTitleCase(s: string) {
  return s.replace(/\w\S*/g, function (txt) {
    return txt.charAt(0).toUpperCase() + txt.substr(1)
  })
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

export function isObject(x: unknown): x is object {
  return Object.prototype.toString.call(x) === '[object Object]'
}

export function getAllSubsets(arr: any[]) {
  return arr.reduce(
    (subsets: any[][], value, index) =>
      subsets.concat(
        subsets.map((set) => {
          const newSet = [...set]
          newSet[index] = value
          return newSet
        })
      ),
    [arr.map(() => undefined)]
  )
}
