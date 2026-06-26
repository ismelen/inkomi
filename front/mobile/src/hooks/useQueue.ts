import { documentDirectory, downloadAsync } from 'expo-file-system/legacy';
import mime from 'mime';
import { create } from 'zustand';
import { copyToCache } from '../../modules/file-handler';
import { BACKENDD_URL } from '../constants';
import { QueueElement } from '../models/queue-element';
import { TransactionRequest } from '../models/transaction-request';
import { FilesystemService } from '../services/filesystem-service';
import { StorageService } from '../services/storage-service';
import { cache } from 'react';

interface State {
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
  transactions: [],
  completedTransactions: [],

  async init() {
    const trans = await StorageService.GetAsync<QueueElement[]>(TRANSACTIONS_KEY);
    const completedTransactions =
      await StorageService.GetAsync<QueueElement[]>(COMPLETE_TRANSACTIONS_KEY);

    set({
      transactions: trans ?? [],
      completedTransactions: completedTransactions ?? [],
    });
  },

  async download(idx: number, id: string): Promise<boolean> {
    const elem = get().completedTransactions[idx];

    try {
      const rutaLocal = `${documentDirectory}${elem.filename}`;

      const downloadResult = await downloadAsync(
        `${BACKENDD_URL}/transaction/download/${id}`,
        rutaLocal
      );

      if (downloadResult.status === 404) {
        alert('File not available');
        return false;
      }

      if (downloadResult.status !== 200) {
        throw Error(`Error al descargar: ${downloadResult.status}`);
      }
    } catch (e: any) {
      alert(e.message);
      return false;
    }

    return true;
  },

  async checkProgress(idx: number, id: string) {
    let progress = 0;
    const elements = get().transactions;
    const elem = elements[idx];

    if (elem.error) return;
    if (elem.progress === 100) return;

    try {
      progress = await fetchStatus(id);
    } catch (e: any) {
      elem.error = e.message;
      return;
    }

    elem.progress = progress;

    if (progress === 100) {
      set({ transactions: [...elements.filter((_, i) => i !== idx)] });
      set((s) => ({ completedTransactions: [elem, ...s.completedTransactions] }));
      StorageService.SetAsync(TRANSACTIONS_KEY, get().transactions);
      StorageService.SetAsync(COMPLETE_TRANSACTIONS_KEY, get().completedTransactions);
      return;
    }

    set({ transactions: [...elements] });
  },

  async send(req: TransactionRequest): Promise<boolean> {
    const isEmptySource = !req.sources || req.sources.length === 0;
    if (isEmptySource) {
      alert('No sources');
      return false;
    }

    const form = new FormData();

    if (req.mode === 'folder') {
      if ((req.sources[0].children?.length ?? 0) === 0) {
        alert('Empty folder');
        return false;
      }
      const formFiles = await Promise.all(
        req.sources[0].children!.map(async (path) => {
          const name = decodeURIComponent(path).split('/').pop();
          if (!name) {
            alert('No file name');
            return false;
          }
          const cachePath = await copyToCache(path, name);
          return {
            name: name,
            type: mime.getType(cachePath),
            uri: cachePath,
          };
        })
      );

      for (const file of formFiles) {
        form.append('files', file as unknown as Blob);
      }
    } else {
      for (let src of req.sources!) {
        form.append(`files`, {
          name: src.name,
          type: mime.getType(src.path),
          uri: src.path,
        } as unknown as Blob);
      }
    }

    const toCloud = (req.destination ?? 'local') === 'cloud';

    form.append('profile', 'KoCC'); //TODO: settings
    form.append('title', req.title ?? '');
    form.append('author', req.author ?? '');
    form.append('cloud', `${toCloud}`);
    form.append('merge', `${req.merge ?? false}`);
    // form.append('notify_token', ''); //TODO: settings

    if (toCloud) {
      form.append('cloud_token', ''); // TODO: cloud service
      form.append('cloud_folder', ''); //TODO: cloud service
    }

    //TODO: a partir de aqui ya es asíncrono, retornar true de momento, guardar en el queue el request para poder lanzarlo de nuevo si falla
    const resp = await fetch(`${BACKENDD_URL}/transaction/convert`, {
      method: 'POST',
      body: form,
    });

    if (resp.status !== 200) {
      const json = await resp.json();
      console.log(json);
      alert(json.error);
      return false;
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
          FilesystemService.deleteFile(child);
        }
        if (!hasChildren) FilesystemService.deleteFile(src.path);
      }
    }

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
