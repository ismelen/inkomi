import { http } from './http-client';
import {
  AddWishRequest,
  AuthorWorksQuery,
  WishItem,
  WishKind,
  WishSuggestion,
  WorksPage,
} from './types/wishlist';

export const wishlistApi = {
  list: () => http.get<WishItem[]>('/wishlist'),
  add: (item: AddWishRequest) => http.post<WishItem>('/wishlist', item),
  remove: (id: number) => http.delete<void>(`/wishlist/${id}`),
  lookup: (kind: WishKind, q: string, signal?: AbortSignal) =>
    http.get<WishSuggestion[]>('/wishlist/lookup', { query: { kind, q }, signal }),
  authorWorks: ({ authorKey, author, sort, lang, page }: AuthorWorksQuery, signal?: AbortSignal) =>
    http.get<WorksPage>('/wishlist/works', {
      query: {
        sort,
        page: String(page),
        ...(authorKey ? { authorKey } : { author: author ?? '' }),
        ...(lang ? { lang } : {}),
      },
      signal,
    }),
};
