import { create } from 'zustand';
import { booksApi } from '../api/books-api';
import { errorKey } from '../api/api-error';
import { BookFormat, BookResponse, SearchField } from '../api/types/books';
import i18n from '../i18n/i18n';
import { BookLanguage, bookLanguages, isBookLanguage } from '../models/book-languages';
import { supportedFormats } from '../models/list-item';
import { useSettings } from './use-settings';

type SearchStatus = 'idle' | 'loading' | 'done' | 'error';

const defaultFormat: BookFormat = 'epub';
const defaultField: SearchField = 'title';

/** Language of the app, the one searched until the user picks another. */
export function defaultSearchLanguage(): BookLanguage | null {
  const lang = useSettings.getState().language ?? i18n.language;
  return isBookLanguage(lang) ? lang : null;
}

interface SearchState {
  query: string;
  field: SearchField;
  format: BookFormat | null;
  /** Language filter; undefined follows the app language, null searches every language. */
  language?: BookLanguage | null;
  results: BookResponse[];
  /** Page to load next; 0 when there are no more results. */
  nextPage: number;
  loadingMore: boolean;
  status: SearchStatus;
  /** i18n key of the last error. */
  error?: string;
  setQuery(query: string): void;
  setField(field: SearchField): void;
  setFormat(format: BookFormat | null): void;
  setLanguage(language: BookLanguage | null): void;
  search(): Promise<void>;
  /** Appends the next page of the current search, if any. */
  loadMore(): Promise<void>;
  reset(): void;
}

let controller: AbortController | null = null;

function fetchPage(state: SearchState, q: string, page: number, signal: AbortSignal) {
  const lang = state.language === undefined ? defaultSearchLanguage() : state.language;
  return booksApi.search(
    q,
    state.field,
    state.format ? [state.format] : supportedFormats,
    lang ? bookLanguages[lang] : undefined,
    page,
    signal
  );
}

export const useSearch = create<SearchState>()((set, get) => ({
  query: '',
  field: defaultField,
  format: defaultFormat,
  results: [],
  nextPage: 0,
  loadingMore: false,
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

  setLanguage: (language) => {
    set({ language });
    void get().search();
  },

  search: async () => {
    const q = get().query.trim();
    controller?.abort();
    if (!q) {
      set({ results: [], nextPage: 0, loadingMore: false, status: 'idle', error: undefined });
      return;
    }

    controller = new AbortController();
    const signal = controller.signal;
    set({ status: 'loading', loadingMore: false, error: undefined });
    try {
      const page = await fetchPage(get(), q, 1, signal);
      if (!signal.aborted) set({ results: page.books, nextPage: page.nextPage, status: 'done' });
    } catch (err) {
      if (!signal.aborted) set({ results: [], nextPage: 0, status: 'error', error: errorKey(err) });
    }
  },

  loadMore: async () => {
    const state = get();
    const q = state.query.trim();
    if (state.status !== 'done' || state.loadingMore || !state.nextPage || !q) return;

    controller = new AbortController();
    const signal = controller.signal;
    set({ loadingMore: true });
    try {
      const page = await fetchPage(state, q, state.nextPage, signal);
      if (signal.aborted) return;
      const seen = new Set(get().results.map((b) => b.md5 || b.title));
      const fresh = page.books.filter((b) => !seen.has(b.md5 || b.title));
      set({ results: [...get().results, ...fresh], nextPage: page.nextPage, loadingMore: false });
    } catch {
      // the list stays as it is; scrolling again retries the same page
      if (!signal.aborted) set({ loadingMore: false });
    }
  },

  reset: () => {
    controller?.abort();
    set({
      query: '',
      field: defaultField,
      format: defaultFormat,
      language: undefined,
      results: [],
      nextPage: 0,
      loadingMore: false,
      status: 'idle',
      error: undefined,
    });
  },
}));
