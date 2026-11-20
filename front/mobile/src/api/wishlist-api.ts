import { http } from './http-client';
import { AddWishRequest, WishItem, WishKind, WishSuggestion } from './types/wishlist';

export const wishlistApi = {
  list: () => http.get<WishItem[]>('/wishlist'),
  add: (item: AddWishRequest) => http.post<WishItem>('/wishlist', item),
  remove: (id: number) => http.delete<void>(`/wishlist/${id}`),
  lookup: (kind: WishKind, q: string, signal?: AbortSignal) =>
    http.get<WishSuggestion[]>('/wishlist/lookup', { query: { kind, q }, signal }),
};
