import { DBSchema, IDBPDatabase, openDB } from 'idb'

import { ExperimentKind } from 'components/NewExperiment/types'

export interface PreDefinedValue {
  name: string
  kind: ExperimentKind
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
