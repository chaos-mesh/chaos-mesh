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
import { DateTime } from 'luxon'

export function comparator(a: string, b: string) {
  const da = DateTime.fromISO(a)
  const db = DateTime.fromISO(b)

  if (da > db) {
    return 1
  }

  if (da < db) {
    return -1
  }

  return 0
}

export const now = DateTime.local

export const format = (date: string, locale: string = 'en') =>
  DateTime.fromISO(date, { locale }).toFormat('yyyy-MM-dd HH:mm:ss a')

export const toRelative = (date: string, locale: string = 'en') => DateTime.fromISO(date, { locale }).toRelative()

export default DateTime
