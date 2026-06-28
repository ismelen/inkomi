import { create } from 'zustand';
import { StorageService } from '../services/storage-service';

interface State {
  model?: string;
  setModel(model?: string): void;
  init(): Promise<void>;
}

const MODEL_KEY = 'e_reader_model';

export const useSettings = create<State>((set, get) => ({
  async init() {
    const model = await StorageService.GetAsync<string>(MODEL_KEY);
    set({ model: model });
  },

  setModel(model?: string) {
    set({ model });
    StorageService.SetAsync(MODEL_KEY, model);
  },
}));
