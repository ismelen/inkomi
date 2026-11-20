import { create } from 'zustand';
import { errorKey } from '../api/api-error';
import { wishlistApi } from '../api/wishlist-api';
import { AddWishRequest, WishItem } from '../api/types/wishlist';

type Status = 'idle' | 'loading' | 'done' | 'error';

interface WishlistState {
  items: WishItem[];
  status: Status;
  /** i18n key of the last error. */
  error?: string;
  load(): Promise<void>;
  /** Adds the item and returns the i18n key of the error, if any. */
  add(item: AddWishRequest): Promise<string | undefined>;
  remove(id: number): Promise<string | undefined>;
}

export const useWishlist = create<WishlistState>()((set, get) => ({
  items: [],
  status: 'idle',

  load: async () => {
    set({ status: 'loading', error: undefined });
    try {
      set({ items: await wishlistApi.list(), status: 'done' });
    } catch (err) {
      set({ status: 'error', error: errorKey(err) });
    }
  },

  add: async (item) => {
    try {
      const added = await wishlistApi.add(item);
      set({ items: [added, ...get().items] });
      return undefined;
    } catch (err) {
      return errorKey(err);
    }
  },

  remove: async (id) => {
    const previous = get().items;
    set({ items: previous.filter((item) => item.id !== id) });
    try {
      await wishlistApi.remove(id);
      return undefined;
    } catch (err) {
      set({ items: previous });
      return errorKey(err);
    }
  },
}));
