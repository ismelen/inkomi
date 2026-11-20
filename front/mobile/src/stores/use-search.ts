import { create } from 'zustand';
import { booksApi } from '../api/books-api';
import { errorKey } from '../api/api-error';
import { BookFormat, BookResponse, SearchField } from '../api/types/books';
import { supportedFormats } from '../models/list-item';

type SearchStatus = 'idle' | 'loading' | 'done' | 'error';

const defaultFormat: BookFormat = 'epub';
const defaultField: SearchField = 'title';

interface SearchState {
  query: string;
  field: SearchField;
  format: BookFormat | null;
  results: BookResponse[];
  status: SearchStatus;
  /** i18n key of the last error. */
  error?: string;
  setQuery(query: string): void;
  setField(field: SearchField): void;
  setFormat(format: BookFormat | null): void;
  search(): Promise<void>;
  reset(): void;
}

let controller: AbortController | null = null;

export const useSearch = create<SearchState>()((set, get) => ({
  query: '',
  field: defaultField,
  format: defaultFormat,
  results: [],
  status: 'idle',

  setQuery: (query) => set({ query }),

  setField: (field) => {
    set({ field });
    void get().search();
  },

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
      const { field, format } = get();
      const results = await booksApi.search(q, field, format ? [format] : supportedFormats, signal);
      if (!signal.aborted) set({ results, status: 'done' });
    } catch (err) {
      if (!signal.aborted) set({ results: [], status: 'error', error: errorKey(err) });
    }
  },

  reset: () => {
    controller?.abort();
    set({
      query: '',
      field: defaultField,
      format: defaultFormat,
      results: [],
      status: 'idle',
      error: undefined,
    });
  },
}));
