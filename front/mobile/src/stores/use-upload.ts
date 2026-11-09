import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import { errorDetail, errorKey } from '../api/api-error';
import { runInBackground } from '../services/background-task';
import { ListSnapshot, UploadJob, UploadRun } from '../services/upload-service';
import i18n from '../i18n/i18n';
import { useList } from './use-list';
import { useSettings } from './use-settings';

export type UploadFileState = 'waiting' | 'uploading' | 'uploaded' | 'failed';

export interface UploadFile {
  sourceId: string;
  title: string;
  progress: number;
  state: UploadFileState;
  /** i18n key. */
  error?: string;
  /** Raw message when the error has no translation. */
  detail?: string;
}

type Phase = 'idle' | 'requesting' | 'uploading' | 'finished' | 'failed';

interface UploadState {
  visible: boolean;
  phase: Phase;
  files: UploadFile[];
  /** i18n key of a request error. */
  error?: string;
  send(toCloud: boolean): Promise<void>;
  /** Uploads again the failed files of the last sends, into their original folders. */
  retryFailed(): Promise<void>;
  /** Whether some list item failed to upload and can be retried. */
  hasFailed(): boolean;
  cancel(): void;
  close(): void;
}

let current: UploadRun | null = null;

/**
 * Jobs whose upload failed, by list item id. Their sources already exist on the server (maybe
 * inside a folder being joined), so sending them again must reuse them instead of creating a new
 * request, or the original folder would wait forever for the missing child.
 */
const failedJobs = new Map<string, UploadJob>();

/** The list without the items that are retried instead of sent as new sources. */
function withoutRetried(list: ListSnapshot, retried: Set<string>): ListSnapshot {
  const items = Object.fromEntries(Object.entries(list.items).filter(([id]) => !retried.has(id)));
  const groups = Object.fromEntries(
    Object.entries(list.groups).map(([id, g]) => [
      id,
      { ...g, itemIds: g.itemIds.filter((itemId) => !retried.has(itemId)) },
    ])
  );
  return { items, groups, rootOrder: list.rootOrder.filter((id) => !retried.has(id)) };
}

function isEmpty(list: ListSnapshot): boolean {
  return list.rootOrder.every((id) => {
    const group = list.groups[id];
    return group ? group.itemIds.length === 0 : !list.items[id];
  });
}

export const useUpload = create<UploadState>()(
  immer((set, get) => {
    const update = (sourceId: string, patch: Partial<UploadFile>) =>
      set((s) => {
        const file = s.files.find((f) => f.sourceId === sourceId);
        if (file) Object.assign(file, patch);
      });

    const newRun = () =>
      new UploadRun({
        onJobs: (jobs) =>
          set((s) => {
            s.phase = 'uploading';
            const known = new Set(s.files.map((f) => f.sourceId));
            for (const j of jobs) {
              if (known.has(j.sourceId)) {
                const file = s.files.find((f) => f.sourceId === j.sourceId)!;
                Object.assign(file, {
                  progress: 0,
                  state: 'waiting',
                  error: undefined,
                  detail: undefined,
                });
              } else {
                s.files.push({
                  sourceId: j.sourceId,
                  title: j.title,
                  progress: 0,
                  state: 'waiting',
                });
              }
            }
          }),
        onProgress: (id, progress) => update(id, { progress, state: 'uploading' }),
        onConfirmed: (job) => {
          failedJobs.delete(job.listItemId);
          update(job.sourceId, { progress: 1, state: 'uploaded' });
        },
        onFailed: (job, err) => {
          console.warn('upload failed', job.sourceId, err);
          failedJobs.set(job.listItemId, job);
          update(job.sourceId, { state: 'failed', error: errorKey(err), detail: errorDetail(err) });
        },
      });

    /** Runs the new items of list and the retries; returns the confirmed list item ids. */
    const execute = async (list: ListSnapshot, retries: UploadJob[], toCloud: boolean) => {
      const { ereaderKey, defaultManga } = useSettings.getState();
      const run = newRun();
      current = run;
      try {
        const confirmed = await runInBackground(
          'Inkomi',
          i18n.t('sending.notification'),
          async () => {
            const ids: string[] = [];
            if (!isEmpty(list)) {
              ids.push(...(await run.run(list, { ereaderKey, toCloud, defaultManga })));
            }
            if (retries.length > 0) ids.push(...(await run.retry(retries)));
            return ids;
          }
        );
        // Sent items leave the list; failed ones stay to be retried.
        useList.getState().remove(confirmed);
        set({ phase: get().files.some((f) => f.state === 'failed') ? 'failed' : 'finished' });
      } catch (err) {
        set({ phase: 'failed', error: errorKey(err) });
      } finally {
        current = null;
      }
    };

    const pendingRetries = () => {
      const items = useList.getState().items;
      // a failed item removed from the list is not retried any more
      for (const id of [...failedJobs.keys()]) if (!items[id]) failedJobs.delete(id);
      return [...failedJobs.values()];
    };

    return {
      visible: false,
      phase: 'idle',
      files: [],

      send: async (toCloud) => {
        if (current) return;
        const { items, groups, rootOrder } = useList.getState();
        const retries = pendingRetries();
        const list = withoutRetried(
          { items, groups, rootOrder },
          new Set(retries.map((j) => j.listItemId))
        );
        set({ visible: true, phase: 'requesting', files: [], error: undefined });
        await execute(list, retries, toCloud);
      },

      retryFailed: async () => {
        if (current) return;
        const retries = pendingRetries();
        if (retries.length === 0) return;
        set({ visible: true, phase: 'uploading', error: undefined });
        await execute({ items: {}, groups: {}, rootOrder: [] }, retries, false);
      },

      hasFailed: () => pendingRetries().length > 0,

      cancel: () => {
        current?.cancel();
        set({ visible: false });
      },

      close: () => set({ visible: false }),
    };
  })
);
