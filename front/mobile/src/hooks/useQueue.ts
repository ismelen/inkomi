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

interface State {
  uploads: Upload[];
  transactions: QueueElement[];
  completedTransactions: QueueElement[];
  send(req: Partial<TransactionRequest>): Promise<boolean>;
  checkProgress(idx: number, id: string): Promise<void>;
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
    // const elem = get().completedTransactions[idx];

    // try {
    //   const dstPath = `${cacheDirectory}${elem.filename}`;

    //   const downloadResult = await downloadAsync(
    //     `${BACKENDD_URL}/transaction/download/${id}`,
    //     dstPath
    //   );

    //   if (downloadResult.status === 404) {
    //     alert('File not available');
    //     return false;
    //   }

    //   if (downloadResult.status !== 200) {
    //     throw Error(`Error al descargar: ${downloadResult.status}`);
    //   }

    //   await shareAsync(dstPath, {
    //     mimeType: 'application/epub+zip',
    //     dialogTitle: 'Guardar archivo',
    //   });
    // } catch (e: any) {
    //   alert(e.message);
    //   return false;
    // }

    // return true;
  },

  async checkProgress(idx: number, id: string) {
    let progress = 0;
    const elements = get().transactions;
    const elem = elements[idx];

    if (!elem) return;

    try {
      progress = await fetchStatus(id);
      elem.progress = progress;
    } catch (e: any) {
      elem.error = e.message;
    }

    if (elem.progress === 100 || elem.error) {
      set({ transactions: [...elements.filter((_, i) => i !== idx)] });
      set((s) => ({ completedTransactions: [elem, ...s.completedTransactions] }));
      StorageService.SetAsync(TRANSACTIONS_KEY, get().transactions);
      StorageService.SetAsync(COMPLETE_TRANSACTIONS_KEY, get().completedTransactions);
      return;
    }

    set({ transactions: [...elements] });
  },

  async send(req: TransactionRequest): Promise<boolean> {
    if (req.mode === 'no-select') return false;
    if (req.sources.length === 0) return false;
    if (req.mode === 'folder' && (req.sources[0].children?.length ?? 0) === 0) return false;

    let files: Source[] = [];
    if (req.mode === 'files') {
      //TODO: Divide by size
      files = req.sources;
    }
    if (req.mode === 'folder') {
      files = req.sources[0].children ?? [];
    }
    if (files.length === 0) return false;

    const form = new FormData();
    const toCloud = req.destination === 'cloud';
    form.append('profile', 'KoCC'); //TODO: settings
    form.append('title', req.title!);
    form.append('author', req.author!);
    form.append('cloud', String(toCloud));
    form.append('merge', String(req.merge));
    // form.append('notify_token', ''); //TODO: settings

    if (toCloud) {
      form.append('cloud_token', ''); // TODO: cloud service
      form.append('cloud_folder', ''); //TODO: cloud service
    }

    for (const file of files) {
      form.append('files', {
        uri: file.path,
        name: file.name,
        type: file.mime ?? 'application/zip',
      } as any);
    }

    const upload: Upload = {
      request: req,
      timestamp: Date.now(),
      id: randomUUID(),
    };
    set((s) => ({ uploads: [...s.uploads, upload] }));

    await NotificationService.requestNotificationPermission();
    await BackgroundService.start(
      async () => {
        try {
          const resp = await fetch(`${BACKENDD_URL}/transaction/convert`, {
            method: 'POST',
            body: form,
          });

          if (resp.status !== 200) {
            const json = await resp.json();
            console.log(json);
            alert(json.error);
          }

          const data: QueueElement[] = await resp.json();
          data.forEach((e) => {
            e.destination = req.destination;
            e.timestamp = Date.now();
          });

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
