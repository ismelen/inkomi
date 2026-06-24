import { create } from 'zustand';
import { QueueElement } from '../models/queue-element';
import { TransactionRequest } from '../models/transaction-request';
import { copyToCache } from '../../modules/file-handler';
import mime from 'mime';
import { BACKENDD_URL } from '../constants';
import { FilesystemService } from '../services/filesystem-service';

interface State {
  transactions: QueueElement[];
  completedTransactions: QueueElement[];
  send(req: Partial<TransactionRequest>): Promise<boolean>;
  checkProgress(idx: number, id: string): Promise<void>;
}

export const useQueue = create<State>((set, get) => ({
  transactions: [
    {
      destination: 'local',
      progress: 47,
      title: 'El héroe de las eras',
      id: 'asdfkjnasdkf',
      sources: [],
    },
    {
      destination: 'cloud',
      progress: 93,
      title: 'El imperio final',
      id: 'adfasdf',
      sources: [],
    },
  ],
  completedTransactions: [
    {
      destination: 'local',
      progress: 100,
      title: 'El héroe de las eras',
      id: 'asdfkjnasdkf',
      sources: [],
    },
    {
      destination: 'cloud',
      progress: 100,
      title: 'El imperio final',
      id: 'asdfkjasdfasdfnasdkf',
      sources: [],
    },
  ],

  async checkProgress(idx: number, id: string) {
    let progress = 0;
    const elements = get().transactions;
    const elem = elements[idx];

    if (elem.error) return;
    if (elem.progress === 100) return;

    try {
      progress = await fetchStatus(id);
    } catch (e: any) {
      elem.error = (e as Error).message;
    }

    if (progress === 100) {
      elem.progress = progress;
      set({ transactions: [...elements.filter((_, i) => i !== idx)] });
      set((s) => ({ completedTransactions: [elem, ...s.completedTransactions] }));
      return;
    }

    set({ transactions: [...elements] });
  },

  async send(req: Partial<TransactionRequest>): Promise<boolean> {
    const isEmptySource = !req.sources || req.sources.length === 0;
    if (isEmptySource) {
      alert('No sources');
      return false;
    }

    const isFolderWithoutChildren =
      req.sources?.length === 1 && (req.sources[0].children ?? []).length === 0;
    if (isFolderWithoutChildren) {
      alert('Folder empty');
      return false;
    }

    const form = new FormData();

    if (req.sources?.length === 1) {
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

    form.append('proifle', 'KoCC'); //TODO: settings
    form.append('title', req.title ?? '');
    form.append('author', req.author ?? '');
    form.append('cloud', `${toCloud}`);
    form.append('merge', `${req.merge ?? false}`);
    form.append('notify_token', ''); //TODO: settings

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
      alert(json.error);
      return false;
    }

    const data: QueueElement = await resp.json();
    data.sources = req.sources!;
    data.destination = req.destination ?? 'local';

    set((s) => ({ transactions: [...s.transactions, data] }));

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
