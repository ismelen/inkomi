// Mirror of back/library-service/internal/infra/api/dtos/book-response.go.

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
