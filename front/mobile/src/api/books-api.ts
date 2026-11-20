import { http } from './http-client';
import { BookFormat, BookResponse, SearchField } from './types/books';

export const booksApi = {
  search: (q: string, by: SearchField, formats: BookFormat[], signal?: AbortSignal) =>
    http.get<BookResponse[]>('/books/search', {
      query: { q, by, fmt: formats.join(',') },
      signal,
    }),
};
