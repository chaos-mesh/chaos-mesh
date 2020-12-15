import { IDBPDatabase, openDB } from 'idb'

let db: IDBPDatabase

export async function getStore(name: string, type: 'readonly' | 'readwrite' = 'readonly') {
  if (!db) {
    db = await openDB('chaos-mesh', 1, {
      upgrade(db, oldVersion) {
        if (oldVersion == 0) {
          db.createObjectStore('predefined', { keyPath: 'id' })
        }
      },
    })
  }

  return db.transaction(name, type).objectStore(name)
}
