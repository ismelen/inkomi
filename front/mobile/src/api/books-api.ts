import { http } from './http-client';
import { BookFormat, SearchField, SearchResponse } from './types/books';

export const booksApi = {
  /** language is the libgen name ("Spanish"); without it every language is searched. page starts at 1. */
  search: (
    q: string,
    by: SearchField,
    formats: BookFormat[],
    language: string | undefined,
    page: number,
    signal?: AbortSignal
  ) =>
    http.get<SearchResponse>('/books/search', {
      query: {
        q,
        by,
        fmt: formats.join(','),
        page: String(page),
        ...(language ? { lang: language } : {}),
      },
      signal,
    }),
};
