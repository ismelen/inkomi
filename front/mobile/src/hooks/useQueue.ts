import { create } from 'zustand';
import { QueueElement } from '../models/queue-element';
import { TransactionRequest } from '../models/transaction-request';
import { copyToCache } from '../../modules/file-handler';
import mime from 'mime';
import { BACKENDD_URL } from '../constants';
import { FilesystemService } from '../services/filesystem-service';
import { documentDirectory, EncodingType, writeAsStringAsync } from 'expo-file-system/legacy';
import { isAvailableAsync, shareAsync } from 'expo-sharing';
import { StorageService } from '../services/storage-service';

interface State {
  transactions: QueueElement[];
  completedTransactions: QueueElement[];
  send(req: Partial<TransactionRequest>): Promise<boolean>;
  checkProgress(idx: number, id: string): Promise<void>;
  download(id: string): Promise<boolean>;
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

  async download(id: string): Promise<boolean> {
    try {
      const resp = await fetch(`${BACKENDD_URL}/transaction/download/${id}`, {
        method: 'GET',
      });

      if (resp.status === 404) {
        alert('File not available');
        return false;
      }

      if (resp.status !== 200) {
        const data = await resp.json();
        throw Error(data.error);
      }

      const contentDisposition = resp.headers.get('content-disposition');
      let nombreArchivo = 'archivo_descargado.dat'; // Nombre por defecto

      if (contentDisposition && contentDisposition.includes('filename=')) {
        const match = contentDisposition.match(/filename=["']?([^"'\s]+)["']?/);
        if (match && match[1]) {
          nombreArchivo = match[1];
        }
      }

      const blob = await resp.blob();

      const reader = new FileReader();
      reader.readAsDataURL(blob);

      reader.onloadend = async () => {
        const resultado = reader.result as string;
        if (!resultado) return;

        const base64data = resultado.split(',')[1];

        const mimeTypeReal = resp.headers.get('content-type') || 'application/octet-stream';

        const rutaLocal = `${documentDirectory}${nombreArchivo}`;
        await writeAsStringAsync(rutaLocal, base64data, {
          encoding: EncodingType.Base64,
        });

        if (await isAvailableAsync()) {
          await shareAsync(rutaLocal, {
            mimeType: mimeTypeReal,
            dialogTitle: 'Guardar archivo',
          });
        } else {
          alert('Error: La función de compartir no está disponible');
        }
      };
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

    const isFolder = req.sources?.length === 1 && (req.sources[0].children ?? []).length !== 0;

    if (isFolder) {
      for (let path of req.sources[0].children!) {
        const name = decodeURIComponent(path).split('/').pop();
        if (!name) {
          alert('No file name');
          return false;
        }
        const cachePath = await copyToCache(path, name);

        form.append('files', {
          name: name,
          type: mime.getType(cachePath),
          uri: cachePath,
        } as unknown as Blob);
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
