import { http } from './http-client';
import { BookFormat, BookResponse, SearchField } from './types/books';

export const booksApi = {
  /** language is the libgen name ("Spanish"); without it every language is searched. */
  search: (
    q: string,
    by: SearchField,
    formats: BookFormat[],
    language: string | undefined,
    signal?: AbortSignal
  ) =>
    http.get<BookResponse[]>('/books/search', {
      query: { q, by, fmt: formats.join(','), ...(language ? { lang: language } : {}) },
      signal,
    }),
};
