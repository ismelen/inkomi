import { create } from 'zustand';
import { UploadTree } from '../models/upload-tree';
import { listHistory, listQueue, onActivityChange } from '../services/activity-db';
import { loadOlderHistory } from '../services/sync-service';

const historyPage = 30;

interface ActivityState {
  queue: UploadTree[];
  history: UploadTree[];
  historyEnd: boolean;
  loadingMore: boolean;
  reload(): Promise<void>;
  loadMoreHistory(): Promise<void>;
}

let reloadScheduled = false;

export const useActivity = create<ActivityState>()((set, get) => ({
  queue: [],
  history: [],
  historyEnd: false,
  loadingMore: false,

  reload: async () => {
    const limit = Math.max(historyPage, get().history.length);
    const [queue, history] = await Promise.all([listQueue(), listHistory(limit)]);
    set({ queue, history });
  },

  /** Next local page; when the local history runs out, the next page from the server. */
  loadMoreHistory: async () => {
    const { history, loadingMore, historyEnd } = get();
    if (loadingMore || historyEnd) return;
    set({ loadingMore: true });
    try {
      const last = history[history.length - 1]?.root;
      const before = last ? { createdAt: last.createdAt, id: last.id } : undefined;
      let page = await listHistory(historyPage, before);
      if (page.length < historyPage) {
        const more = await loadOlderHistory();
        page = await listHistory(historyPage, before);
        if (!more && page.length < historyPage) set({ historyEnd: true });
      }
      const known = new Set(get().history.map((t) => t.root.id));
      set({ history: [...get().history, ...page.filter((t) => !known.has(t.root.id))] });
    } catch (err) {
      console.warn('load history', err);
    } finally {
      set({ loadingMore: false });
    }
  },
}));

// Database changes come in bursts (one per event): reload once per tick.
onActivityChange(() => {
  if (reloadScheduled) return;
  reloadScheduled = true;
  setTimeout(() => {
    reloadScheduled = false;
    void useActivity.getState().reload();
  }, 50);
});
