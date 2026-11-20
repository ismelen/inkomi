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
  status: SearchStatus;
  /** i18n key of the last error. */
  error?: string;
  setQuery(query: string): void;
  setField(field: SearchField): void;
  setFormat(format: BookFormat | null): void;
  setLanguage(language: BookLanguage | null): void;
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

  setLanguage: (language) => {
    set({ language });
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
      const { field, format, language } = get();
      const lang = language === undefined ? defaultSearchLanguage() : language;
      const results = await booksApi.search(
        q,
        field,
        format ? [format] : supportedFormats,
        lang ? bookLanguages[lang] : undefined,
        signal
      );
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
      language: undefined,
      results: [],
      status: 'idle',
      error: undefined,
    });
  },
}));
