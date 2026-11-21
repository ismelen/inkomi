import { useCallback, useEffect, useRef, useState } from 'react';
import { wishlistApi } from '../api/wishlist-api';
import { WishKind, WishSuggestion } from '../api/types/wishlist';

const debounceMs = 400;
const minQuery = 2;

/** Open Library suggestions for a book or author name, debounced; a new lookup aborts the last. */
export function useCatalogSuggestions() {
  const [suggestions, setSuggestions] = useState<WishSuggestion[]>([]);
  const [loading, setLoading] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const controller = useRef<AbortController | null>(null);

  const cancel = useCallback(() => {
    if (timer.current) clearTimeout(timer.current);
    controller.current?.abort();
  }, []);

  const clear = useCallback(() => {
    cancel();
    setSuggestions([]);
    setLoading(false);
  }, [cancel]);

  const lookup = useCallback(
    (kind: WishKind, text: string) => {
      cancel();
      const q = text.trim();
      if (q.length < minQuery) {
        setSuggestions([]);
        setLoading(false);
        return;
      }
      timer.current = setTimeout(() => {
        const ctrl = new AbortController();
        controller.current = ctrl;
        setLoading(true);
        wishlistApi
          .lookup(kind, q, ctrl.signal)
          .then((found) => !ctrl.signal.aborted && setSuggestions(found))
          .catch(() => !ctrl.signal.aborted && setSuggestions([]))
          .finally(() => !ctrl.signal.aborted && setLoading(false));
      }, debounceMs);
    },
    [cancel]
  );

  useEffect(() => cancel, [cancel]);

  return { suggestions, loading, lookup, clear };
}
