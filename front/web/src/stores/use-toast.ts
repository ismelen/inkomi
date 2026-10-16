import { create } from 'zustand';
import { errorKey } from '../api/api-error';
import i18n from '../i18n/i18n';

type Tone = 'neutral' | 'danger';

interface Toast {
  id: number;
  message: string;
  tone: Tone;
}

interface ToastState {
  toasts: Toast[];
  show(message: string, tone?: Tone): void;
  dismiss(id: number): void;
}

let nextId = 1;
const visibleMs = 5_000;

/** Short messages instead of the app's alerts. */
export const useToast = create<ToastState>()((set, get) => ({
  toasts: [],
  show: (message, tone = 'neutral') => {
    const id = nextId++;
    set({ toasts: [...get().toasts.slice(-2), { id, message, tone }] });
    setTimeout(() => get().dismiss(id), visibleMs);
  },
  dismiss: (id) => set({ toasts: get().toasts.filter((t) => t.id !== id) }),
}));

export function showError(err: unknown) {
  useToast.getState().show(i18n.t(errorKey(err)), 'danger');
}
