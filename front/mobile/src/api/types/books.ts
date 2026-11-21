// Mirror of back/api/internal/modules/library/ui/dtos/book-response.go and search-response.go.

export type BookFormat = 'epub' | 'cbz' | 'pdf';

/** Book field the search query is matched against (`by` of /books/search). */
export type SearchField = 'title' | 'author';

export interface BookResponse {
  title: string;
  author: string;
  pages: number;
  language: string;
  extension: string;
  /** Empty when the mirror didn't expose it: the book can't be added. */
  md5: string;
  /** Path of the api (/books/cover?url=...) that serves the mirror's cover; empty without one. */
  coverUrl: string;
}

/** A page of GET /books/search. */
export interface SearchResponse {
  books: BookResponse[];
  /** Page to ask for to see more books; 0 when there are no more. */
  nextPage: number;
}
