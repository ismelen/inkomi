import { useSyncExternalStore } from 'react';

/** Tailwind's lg breakpoint: the layout of the "Web" frame. */
export const desktopQuery = '(min-width: 1024px)';

/** Below Tailwind's md: phones, with the app's navigation (no sidebar). */
export const phoneQuery = '(max-width: 767px)';

export function useMediaQuery(query: string): boolean {
  return useSyncExternalStore(
    (onChange) => {
      const media = window.matchMedia(query);
      media.addEventListener('change', onChange);
      return () => media.removeEventListener('change', onChange);
    },
    () => window.matchMedia(query).matches
  );
}
