import { create } from 'zustand';
import { BACKENDD_URL } from '../constants';
import { LibgenBook } from '../models/book';

interface State {
  search(query?: string): Promise<LibgenBook[] | undefined>;
  selected: Record<string, LibgenBook>;

  selectBook(book: LibgenBook): void;
  clear(): void;
}

export const useLibgen = create<State>((set, get) => ({
  selected: {},

  async search(query?: string): Promise<LibgenBook[] | undefined> {
    if (!query) return;

    const resp = await fetch(`${BACKENDD_URL}/books/search?q=${query ?? ''}`, {
      method: 'GET',
    });

    if (!resp.ok || resp.status !== 200) return;
    return await resp.json();
  },

  clear() {
    set({ selected: {} });
  },

  selectBook(book: LibgenBook) {
    const exists = !!get().selected[book.md5];
    const selected = get().selected;

    if (exists) {
      delete selected[book.md5];
    } else {
      selected[book.md5] = book;
    }

    set({ selected: { ...selected } });
  },
}));
