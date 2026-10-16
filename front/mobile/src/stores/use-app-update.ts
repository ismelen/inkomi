import { create } from 'zustand';
import { errorKey } from '../api/api-error';
import { ReleaseResponse } from '../api/types/app';
import { findUpdate, installUpdate } from '../services/app-update';

type Phase = 'idle' | 'available' | 'downloading' | 'error';

interface AppUpdateState {
  release?: ReleaseResponse;
  phase: Phase;
  progress: number;
  /** i18n key. */
  error?: string;
  /** Dismissed with "later": not shown again until the next launch. */
  dismissed: boolean;
  lastCheck: number;
  /** Returns whether there is an update. manual ignores the dismiss and throws on errors. */
  check(manual?: boolean): Promise<boolean>;
  install(): Promise<void>;
  dismiss(): void;
}

export const useAppUpdate = create<AppUpdateState>()((set, get) => ({
  phase: 'idle',
  progress: 0,
  dismissed: false,
  lastCheck: 0,

  check: async (manual = false) => {
    if (get().phase === 'downloading') return true;
    set({ lastCheck: Date.now() });
    try {
      const release = await findUpdate();
      if (!release) {
        set({ release: undefined, phase: 'idle' });
        return false;
      }
      if (manual || !get().dismissed) set({ release, phase: 'available', dismissed: false });
      return true;
    } catch (err) {
      if (manual) throw err;
      return false;
    }
  },

  install: async () => {
    const release = get().release;
    if (!release) return;
    set({ phase: 'downloading', progress: 0, error: undefined });
    try {
      await installUpdate(release, (progress) => set({ progress }));
      // The installer is open; if the user cancels it, the dialog lets them try again.
      set({ phase: 'available' });
    } catch (err) {
      set({ phase: 'error', error: errorKey(err) });
    }
  },

  dismiss: () => set({ phase: 'idle', dismissed: true }),
}));
