import { create } from 'zustand';
import { booksApi } from '../api/books-api';
import { errorKey } from '../api/api-error';
import { BookFormat, BookResponse } from '../api/types/books';
import { supportedFormats } from '../models/list-item';

type SearchStatus = 'idle' | 'loading' | 'done' | 'error';

interface SearchState {
  query: string;
  format: BookFormat | null;
  results: BookResponse[];
  status: SearchStatus;
  /** i18n key of the last error. */
  error?: string;
  setQuery(query: string): void;
  setFormat(format: BookFormat | null): void;
  search(): Promise<void>;
  reset(): void;
}

let controller: AbortController | null = null;

export const useSearch = create<SearchState>()((set, get) => ({
  query: '',
  format: null,
  results: [],
  status: 'idle',

  setQuery: (query) => set({ query }),

  setFormat: (format) => {
    set({ format });
    void get().search();
  },

  search: async () => {
    const q = get().query.trim();
    controller?.abort();
    if (!q) {
      set({ results: [], status: 'idle', error: undefined });
      return;
    }

    controller = new AbortController();
    const signal = controller.signal;
    set({ status: 'loading', error: undefined });
    try {
      const format = get().format;
      const results = await booksApi.search(q, format ? [format] : supportedFormats, signal);
      if (!signal.aborted) set({ results, status: 'done' });
    } catch (err) {
      if (!signal.aborted) set({ results: [], status: 'error', error: errorKey(err) });
    }
  },

  reset: () => {
    controller?.abort();
    set({ query: '', results: [], status: 'idle', error: undefined });
  },
}));
