/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

const PREFIX = 'chaos-mesh-'

export default class LocalStorage {
  static ls = window.localStorage

  static get(key: string) {
    return this.ls.getItem(PREFIX + key)
  }

  static getObj(key: string) {
    return JSON.parse(this.get(key) ?? '{}')
  }

  static set(key: string, val: string) {
    this.ls.setItem(PREFIX + key, val)
  }

  static setObj(key: string, obj: any) {
    this.set(key, JSON.stringify(obj))
  }

  static remove(key: string) {
    this.ls.removeItem(PREFIX + key)
  }
}
