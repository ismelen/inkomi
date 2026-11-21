// Mirror of back/api/internal/modules/library/ui/dtos/book-response.go and search-response.go.

export type BookFormat = 'epub' | 'cbz' | 'pdf';

export interface BookResponse {
  title: string;
  author: string;
  pages: number;
  language: string;
  extension: string;
  /** Empty when the mirror didn't expose it: the book can't be added. */
  md5: string;
  coverUrl: string;
}

/** A page of GET /books/search. */
export interface SearchResponse {
  books: BookResponse[];
  /** Page to ask for to see more books; 0 when there are no more. */
  nextPage: number;
}
