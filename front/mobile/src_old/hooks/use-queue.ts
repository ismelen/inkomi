import { create } from 'zustand';
import { QueueRequest } from '../models/queue';
import { DownloadService } from '../services/download-service';
import { StorageService } from '../services/storage-service';

const QUEUE = 'queue_key';

interface State {
  requests: QueueRequest[];
  add(request: QueueRequest): void;
  delete(requestIdx: number, timeIdx: number): void;
  download(requestIdx: number, timeIdx: number): Promise<void>;
  init(): Promise<void>;
}

export const useQueue = create<State>((set, get) => ({
  requests: [],

  async init() {
    const requests = await StorageService.GetAsync<QueueRequest[]>(QUEUE);
    if (!requests) return;

    set({ requests: requests });
  },

  add(request: QueueRequest) {
    let requests = get().requests;
    requests = [request, ...requests];

    StorageService.SetAsync(QUEUE, requests);
    set({ requests: requests });
  },

  delete(requestIdx: number, timeIdx: number) {
    let requests = get().requests;

    const request = requests[requestIdx];
    request.times = request.times.filter((_, i) => i !== timeIdx);
    requests[requestIdx] = request;

    if (request.times.length === 0) {
      requests = requests.filter((_, i) => i !== requestIdx);
    }

    StorageService.SetAsync(QUEUE, requests);
    set({ requests: requests });
  },

  async download(requestIdx: number, timeIdx: number) {
    const requests = get().requests;
    const request = requests[requestIdx];
    if (request.settings.cloud) return;

    const time = request.times[timeIdx];

    const now = new Date();
    if ((time.endTime?.getTime() ?? 0) > now.getTime()) return;

    if (!(await DownloadService.download(time.path))) return;
    request.times[timeIdx].endTime = undefined;
    requests[requestIdx] = request;

    // set({ requests: requests });
    get().delete(requestIdx, timeIdx);
  },
}));
