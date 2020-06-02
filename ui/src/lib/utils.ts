export function upperFirst(s: string) {
  if (!s) return ''

  return s.charAt(0).toUpperCase() + s.slice(1)
}
