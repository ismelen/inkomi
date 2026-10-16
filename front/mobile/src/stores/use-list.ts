import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import { randomUUID } from 'expo-crypto';
import { BookResponse } from '../api/types/books';
import { Group } from '../models/group';
import { formatOf, ListItem, titleFromFilename } from '../models/list-item';
import { PickedFile, PickedFolder } from '../services/file-picker';
import { asyncStorage } from './persist-storage';

type ItemPatch = Partial<Pick<ListItem, 'title' | 'rtl' | 'config'>>;
type GroupPatch = Partial<Pick<Group, 'title' | 'join' | 'rtl' | 'config' | 'collapsed'>>;

interface ListState {
  items: Record<string, ListItem>;
  groups: Record<string, Group>;
  /** Loose item ids and group ids, in display order. */
  rootOrder: string[];
  selecting: boolean;
  /** Selected loose item, grouped item or group ids. */
  selection: string[];

  addFiles(files: PickedFile[]): void;
  addFolder(folder: PickedFolder): void;
  /** Adds the book or removes it if it was already in the list. Returns whether it is now added. */
  toggleLibrary(book: BookResponse): boolean;
  updateItem(id: string, patch: ItemPatch): void;
  updateGroup(id: string, patch: GroupPatch): void;
  reorderGroup(id: string, itemIds: string[]): void;
  createGroup(title: string): string;
  /** Removes items and groups (with their items) from the list; files on disk are untouched. */
  remove(ids: string[]): void;
  /** Moves items, and the items of selected groups, into a group or out of any (null). */
  move(ids: string[], groupId: string | null): void;
  startSelection(id: string): void;
  toggleSelected(id: string): void;
  toggleSelectAll(): void;
  clearSelection(): void;
  clear(): void;
}

function fileItem(file: PickedFile, groupId?: string): ListItem {
  return {
    id: randomUUID(),
    kind: 'file',
    title: titleFromFilename(file.name),
    ext: file.ext,
    rtl: false,
    uri: file.uri,
    name: file.name,
    size: file.size,
    mime: file.mime,
    groupId,
  };
}

function removeFromParents(state: ListState, itemId: string) {
  const item = state.items[itemId];
  if (!item) return;
  if (item.groupId && state.groups[item.groupId]) {
    const group = state.groups[item.groupId];
    group.itemIds = group.itemIds.filter((id) => id !== itemId);
  } else {
    state.rootOrder = state.rootOrder.filter((id) => id !== itemId);
  }
}

function dropEmptyGroups(state: ListState) {
  Object.values(state.groups).forEach((group) => {
    if (group.itemIds.length > 0) return;
    delete state.groups[group.id];
    state.rootOrder = state.rootOrder.filter((id) => id !== group.id);
  });
}

/** Item ids behind a selection: groups expand to their items. */
function expand(state: ListState, ids: string[]): string[] {
  const result = new Set<string>();
  ids.forEach((id) => {
    if (state.groups[id]) state.groups[id].itemIds.forEach((itemId) => result.add(itemId));
    else if (state.items[id]) result.add(id);
  });
  return [...result];
}

export const useList = create<ListState>()(
  persist(
    immer((set, get) => ({
      items: {},
      groups: {},
      rootOrder: [],
      selecting: false,
      selection: [],

      addFiles: (files) =>
        set((s) => {
          files.forEach((file) => {
            const item = fileItem(file);
            s.items[item.id] = item;
            s.rootOrder.push(item.id);
          });
        }),

      addFolder: (folder) =>
        set((s) => {
          if (folder.files.length === 0) return;
          const groupId = randomUUID();
          const items = folder.files.map((file) => fileItem(file, groupId));
          items.forEach((item) => (s.items[item.id] = item));
          s.groups[groupId] = {
            id: groupId,
            title: folder.name,
            join: false,
            rtl: false,
            itemIds: items.map((i) => i.id),
            collapsed: false,
          };
          s.rootOrder.push(groupId);
        }),

      toggleLibrary: (book) => {
        const existing = Object.values(get().items).find(
          (i) => i.kind === 'library' && i.md5 === book.md5
        );
        if (existing) {
          get().remove([existing.id]);
          return false;
        }

        const ext = formatOf(`x.${book.extension}`) ?? 'epub';
        set((s) => {
          const id = randomUUID();
          s.items[id] = {
            id,
            kind: 'library',
            title: book.title,
            author: book.author,
            ext,
            rtl: false,
            md5: book.md5,
          };
          s.rootOrder.push(id);
        });
        return true;
      },

      updateItem: (id, patch) =>
        set((s) => {
          if (s.items[id]) Object.assign(s.items[id], patch);
        }),

      updateGroup: (id, patch) =>
        set((s) => {
          if (s.groups[id]) Object.assign(s.groups[id], patch);
        }),

      reorderGroup: (id, itemIds) =>
        set((s) => {
          if (s.groups[id]) s.groups[id].itemIds = itemIds;
        }),

      createGroup: (title) => {
        const id = randomUUID();
        set((s) => {
          s.groups[id] = { id, title, join: false, rtl: false, itemIds: [], collapsed: false };
          s.rootOrder.push(id);
        });
        return id;
      },

      remove: (ids) =>
        set((s) => {
          expand(s, ids).forEach((itemId) => {
            removeFromParents(s, itemId);
            delete s.items[itemId];
          });
          dropEmptyGroups(s);
          s.selection = s.selection.filter((id) => s.items[id] || s.groups[id]);
          if (s.selection.length === 0) s.selecting = false;
        }),

      move: (ids, groupId) =>
        set((s) => {
          const target = groupId ? s.groups[groupId] : null;
          expand(s, ids).forEach((itemId) => {
            const item = s.items[itemId];
            if (item.groupId === (groupId ?? undefined)) return;
            removeFromParents(s, itemId);
            if (target) {
              item.groupId = target.id;
              target.itemIds.push(itemId);
            } else {
              item.groupId = undefined;
              s.rootOrder.push(itemId);
            }
          });
          dropEmptyGroups(s);
          s.selection = [];
          s.selecting = false;
        }),

      startSelection: (id) =>
        set((s) => {
          s.selecting = true;
          s.selection = [id];
        }),

      toggleSelected: (id) =>
        set((s) => {
          s.selection = s.selection.includes(id)
            ? s.selection.filter((x) => x !== id)
            : [...s.selection, id];
          if (s.selection.length === 0) s.selecting = false;
        }),

      toggleSelectAll: () =>
        set((s) => {
          const all = s.rootOrder;
          s.selection = s.selection.length >= all.length ? [] : [...all];
          if (s.selection.length === 0) s.selecting = false;
        }),

      clearSelection: () => set({ selection: [], selecting: false }),

      clear: () => set({ items: {}, groups: {}, rootOrder: [], selection: [], selecting: false }),
    })),
    {
      name: 'list-draft',
      storage: asyncStorage,
      version: 1,
      partialize: (s) => ({ items: s.items, groups: s.groups, rootOrder: s.rootOrder }),
    }
  )
);

export function isLibraryAdded(items: Record<string, ListItem>, md5: string): boolean {
  return Object.values(items).some((i) => i.kind === 'library' && i.md5 === md5);
}
