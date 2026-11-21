import { http } from './http-client';
import { BookFormat, SearchResponse } from './types/books';

export const booksApi = {
  search: (q: string, formats: BookFormat[], signal?: AbortSignal) =>
    http.get<SearchResponse>('/books/search', {
      query: { q, fmt: formats.join(',') },
      signal,
    }),
};
