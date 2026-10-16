import { openDB } from 'idb';
import { createJSONStorage, PersistStorage, StorageValue } from 'zustand/middleware';

/** Small preferences: localStorage, readable before React starts (the theme in index.html). */
export const localJsonStorage = createJSONStorage(() => localStorage);

const db = () =>
  openDB('inkomi-drafts', 1, {
    upgrade(conn) {
      conn.createObjectStore('drafts');
    },
  });

/** IndexedDB without JSON: the values are structured-cloned, so File objects survive. */
export function idbStorage<S>(): PersistStorage<S> {
  return {
    getItem: async (name) => ((await (await db()).get('drafts', name)) as StorageValue<S>) ?? null,
    setItem: async (name, value) => {
      await (await db()).put('drafts', value, name);
    },
    removeItem: async (name) => {
      await (await db()).delete('drafts', name);
    },
  };
}
