import { cacheDirectory, documentDirectory, downloadAsync } from 'expo-file-system/legacy';
import { create } from 'zustand';
import { BACKENDD_URL } from '../constants';
import { QueueElement } from '../models/queue-element';
import { TransactionRequest } from '../models/transaction-request';
import { StorageService } from '../services/storage-service';
import { Source } from '../models/source';
import BackgroundService from 'react-native-background-actions';
import { FilesystemService } from '../services/filesystem-service';
import { NotificationService } from '../services/notification-service';
import { Upload } from '../models/upload';
import { randomUUID } from 'expo-crypto';
import RNBlobUtil from 'react-native-blob-util';
import { useCloud } from './useCloud';
import { LibgenTransactionRequest } from '../models/libgen-transaction-request';
import { LinearGradient } from 'react-native-svg';

interface State {
  uploads: Upload[];
  transactions: QueueElement[];
  completedTransactions: QueueElement[];
  send(req: Partial<TransactionRequest>, libgenMode?: boolean): Promise<boolean>;
  checkProgress(idx: number, id: string): Promise<boolean>;
  download(idx: number, id: string): Promise<boolean>;
  init(): Promise<void>;
}

const TRANSACTIONS_KEY = 'transactions';
const COMPLETE_TRANSACTIONS_KEY = 'complete_transactions';

export const useQueue = create<State>((set, get) => ({
  uploads: [],
  transactions: [],
  completedTransactions: [],

  async init() {
    const trans = (await StorageService.GetAsync<QueueElement[]>(TRANSACTIONS_KEY)) ?? [];
    let completedTransactions =
      (await StorageService.GetAsync<QueueElement[]>(COMPLETE_TRANSACTIONS_KEY)) ?? [];

    if (completedTransactions?.length > 10) {
      completedTransactions = completedTransactions.slice(0, 10);
      StorageService.SetAsync(COMPLETE_TRANSACTIONS_KEY, completedTransactions);
    }

    set({
      transactions: trans,
      completedTransactions: completedTransactions,
    });
  },

  async download(idx: number, id: string): Promise<boolean> {
    const elem = get().completedTransactions[idx];

    try {
      await RNBlobUtil.config({
        addAndroidDownloads: {
          useDownloadManager: true,
          notification: true,
          title: elem.filename,
          description: 'Descargando archivo...',
          mime: 'application/epub+zip',
          mediaScannable: true,
          path: `${RNBlobUtil.fs.dirs.DownloadDir}/${elem.filename}`,
        },
      }).fetch('GET', `${BACKENDD_URL}/transaction/download/${id}`);
    } catch (e: any) {
      alert(e.message);
      return false;
    }

    return true;
  },

  async checkProgress(idx: number, id: string): Promise<boolean> {
    let progress = 0;
    const elem = get().transactions.find((e) => e.id === id);

    if (!elem) return true; // ya fue procesado, detener el intervalo

    try {
      progress = await fetchStatus(id);
      elem.progress = progress;
    } catch (e: any) {
      elem.error = e.message;
    }

    if (elem.progress === 100 || elem.error) {
      set((s) => ({
        transactions: s.transactions.filter((e) => e.id !== id),
        completedTransactions: [elem, ...s.completedTransactions],
      }));
      StorageService.SetAsync(TRANSACTIONS_KEY, get().transactions);
      StorageService.SetAsync(COMPLETE_TRANSACTIONS_KEY, get().completedTransactions);
      return true; // completado, detener el intervalo
    }

    return false;
  },

  async send(req: TransactionRequest, libgenMode?: boolean): Promise<boolean> {
    if (!(libgenMode ?? false)) {
      if (req.mode === 'no-select') return false;
      if (req.sources.length === 0) return false;
      if (req.mode === 'folder' && (req.sources[0].children?.length ?? 0) === 0) return false;
    }
    const form = new FormData();

    if (libgenMode ?? false) {
      const books = (req as LibgenTransactionRequest).books;
      form.append('md5s', books.map((e) => e.md5).join(','));
    } else {
      let files: Source[] = [];
      if (req.mode === 'files') {
        //TODO: Divide by size
        files = req.sources;
      }
      if (req.mode === 'folder') {
        files = req.sources[0].children ?? [];
      }
      if (files.length === 0) return false;

      for (const file of files) {
        form.append('files', {
          uri: file.path,
          name: file.name,
          type: file.mime ?? 'application/zip',
        } as any);
      }
    }

    const toCloud = req.destination === 'cloud';

    form.append('profile', 'KoCC'); //TODO: settings
    form.append('title', req.title ?? '');
    form.append('author', req.author ?? '');
    form.append('cloud', String(toCloud));
    form.append('merge', String(req.merge));
    // form.append('notify_token', ''); //TODO: settings

    if (toCloud) {
      const token = (await useCloud.getState().getToken()) ?? '';
      const folder = (await useCloud.getState().getFolder()) ?? '';
      form.append('cloud_token', token);
      form.append('cloud_folder', folder);
    }

    const upload: Upload = {
      request: req,
      timestamp: Date.now(),
      id: randomUUID(),
      libgenMode: libgenMode ?? false,
    };
    set((s) => ({ uploads: [...s.uploads, upload] }));

    await NotificationService.requestNotificationPermission();
    await BackgroundService.start(
      async () => {
        try {
          const resp = await fetch(
            `${BACKENDD_URL}/transaction/convert${(libgenMode ?? false) ? '?remote=true' : ''}`,
            {
              method: 'POST',
              body: form,
            }
          );

          if (resp.status !== 200) {
            const json = await resp.json();
            alert(json.error);
            return;
          }

          const data: QueueElement[] = await resp.json();
          data.forEach((e) => {
            e.destination = req.destination;
            e.timestamp = Date.now();
          });

          if (libgenMode) {
            const books = (req as LibgenTransactionRequest).books;
            data.forEach((e) => {
              e.title = books.find((b) => b.md5 === e.filename)?.title ?? e.filename;
            });
          }

          set((s) => ({ transactions: [...s.transactions, ...data] }));
          StorageService.SetAsync(TRANSACTIONS_KEY, get().transactions);

          if (req.deleteOrigin ?? false) {
            for (let src of req.sources!) {
              const hasChildren = (src.children ?? []).length > 0;
              for (let child of src.children ?? []) {
                FilesystemService.deleteFile(child.path);
              }
              if (!hasChildren) FilesystemService.deleteFile(src.path);
            }
          }

          set((s) => ({ uploads: s.uploads.filter((e) => e.id !== upload.id) }));
        } catch (e) {
          set((s) => ({
            uploads: s.uploads.map((u) => (u.id === upload.id ? { ...u, error: e as Error } : u)),
          }));
        } finally {
          await BackgroundService.stop();
        }
      },
      {
        taskName: 'inkomi-upload',
        taskTitle: 'Inkomi',
        taskDesc: 'Uploading files...',
        taskIcon: {
          name: 'ic_launcher',
          type: 'mipmap',
        },
        foregroundServiceType: ['dataSync'],
      }
    );

    return true;
  },
}));

async function fetchStatus(id: string): Promise<number> {
  const resp = await fetch(`${BACKENDD_URL}/transaction/status/${id}`, { method: 'GET' });

  const json = await resp.json();

  if (resp.status !== 200) {
    throw new Error(json.error);
  }

  return json.progress;
}
