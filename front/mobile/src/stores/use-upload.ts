import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import { errorKey } from '../api/api-error';
import { runInBackground } from '../services/background-task';
import { UploadRun } from '../services/upload-service';
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
}

type Phase = 'idle' | 'requesting' | 'uploading' | 'finished' | 'failed';

interface UploadState {
  visible: boolean;
  phase: Phase;
  files: UploadFile[];
  /** i18n key of a request error. */
  error?: string;
  send(toCloud: boolean): Promise<void>;
  cancel(): void;
  close(): void;
}

let current: UploadRun | null = null;

export const useUpload = create<UploadState>()(
  immer((set, get) => {
    const update = (sourceId: string, patch: Partial<UploadFile>) =>
      set((s) => {
        const file = s.files.find((f) => f.sourceId === sourceId);
        if (file) Object.assign(file, patch);
      });

    return {
      visible: false,
      phase: 'idle',
      files: [],

      send: async (toCloud) => {
        if (current) return;
        const { items, groups, rootOrder } = useList.getState();
        const { ereaderKey, defaultManga } = useSettings.getState();
        set({ visible: true, phase: 'requesting', files: [], error: undefined });

        const run = new UploadRun({
          onJobs: (jobs) =>
            set((s) => {
              s.phase = 'uploading';
              s.files = jobs.map((j) => ({
                sourceId: j.sourceId,
                title: j.title,
                progress: 0,
                state: 'waiting',
              }));
            }),
          onProgress: (id, progress) => update(id, { progress, state: 'uploading' }),
          onConfirmed: (job) => update(job.sourceId, { progress: 1, state: 'uploaded' }),
          onFailed: (job, err) => update(job.sourceId, { state: 'failed', error: errorKey(err) }),
        });
        current = run;

        try {
          const confirmed = await runInBackground('Inkomi', i18n.t('sending.notification'), () =>
            run.run({ items, groups, rootOrder }, { ereaderKey, toCloud, defaultManga })
          );
          // Sent items leave the list; failed ones stay to be sent again.
          useList.getState().remove(confirmed);
          set({ phase: get().files.some((f) => f.state === 'failed') ? 'failed' : 'finished' });
        } catch (err) {
          set({ phase: 'failed', error: errorKey(err) });
        } finally {
          current = null;
        }
      },

      cancel: () => {
        current?.cancel();
        set({ visible: false });
      },

      close: () => set({ visible: false }),
    };
  })
);
