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

import { DBSchema, IDBPDatabase, openDB } from 'idb'

import { ExperimentKind } from 'components/NewExperiment/types'

export interface PreDefinedValue {
  name: string
  kind: ExperimentKind | 'Schedule'
  yaml: object
}

interface DB extends DBSchema {
  predefined: {
    key: string
    value: PreDefinedValue
  }
}

let db: IDBPDatabase<DB>

export async function getDB() {
  if (!db) {
    db = await openDB<DB>('chaos-mesh', 1, {
      upgrade(db, oldVersion) {
        if (oldVersion === 0) {
          db.createObjectStore('predefined', {
            keyPath: 'name',
          })
        }
      },
    })
  }

  return db
}
