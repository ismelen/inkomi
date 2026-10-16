import { http } from './http-client';
import { BookFormat, BookResponse } from './types/books';

export const booksApi = {
  search: (q: string, formats: BookFormat[], signal?: AbortSignal) =>
    http.get<BookResponse[]>('/books/search', {
      query: { q, fmt: formats.join(',') },
      signal,
    }),
};
