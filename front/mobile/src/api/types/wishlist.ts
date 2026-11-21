// Mirror of back/api/internal/modules/wishlist/ui/dtos.

export type WishKind = 'book' | 'author';

export interface WishItem {
  id: number;
  kind: WishKind;
  /** Book title, or the author's name for kind author. */
  title: string;
  author: string;
  /** Open Library keys; empty when the item was typed by hand. */
  workKey: string;
  authorKey: string;
  createdAt: string;
}

export interface AddWishRequest {
  kind: WishKind;
  title: string;
  author?: string;
  workKey?: string;
  authorKey?: string;
}

/** A catalog match while typing a wish (GET /wishlist/lookup). */
export interface WishSuggestion {
  kind: WishKind;
  title: string;
  /** The book's authors, or the best known work of an author. */
  author: string;
  workKey: string;
  authorKey: string;
  year: number;
  coverUrl: string;
}

/** Order of the works of an author (sort of GET /wishlist/works). */
export type WorksSort = 'popular' | 'new' | 'old' | 'title';

/** A page of the works of an author (GET /wishlist/works). */
export interface WorksPage {
  works: WishSuggestion[];
  total: number;
  page: number;
  hasMore: boolean;
}

export interface AuthorWorksQuery {
  /** Open Library key; without it the author is searched by name. */
  authorKey?: string;
  author?: string;
  sort: WorksSort;
  /** ISO 639-1 code; titles and covers come from the edition in that language. */
  lang?: string;
  page: number;
}
