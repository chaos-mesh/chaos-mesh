const PREFIX = 'chaos-mesh-'

export default class LocalStorage {
  static ls = window.localStorage

  static get(key: string) {
    return LocalStorage.ls.getItem(PREFIX + key)
  }

  static set(key: string, val: string) {
    return LocalStorage.ls.setItem(PREFIX + key, val)
  }

  static remove(key: string) {
    return LocalStorage.ls.removeItem(PREFIX + key)
  }
}
