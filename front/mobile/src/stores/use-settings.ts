import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import i18n, { SupportedLanguage } from '../i18n/i18n';
import { defaultEReaderKey } from '../models/ereaders';
import { defaultMangaConfig, MangaConfig } from '../models/manga-config';
import { asyncStorage } from './persist-storage';

interface SettingsState {
  ereaderKey: string;
  /** Initial value of the Dropbox switch in Home. */
  toCloudDefault: boolean;
  language?: SupportedLanguage;
  defaultManga: MangaConfig;
  notifications: boolean;
  setEReader(key: string): void;
  setToCloudDefault(value: boolean): void;
  setLanguage(language: SupportedLanguage): Promise<void>;
  setDefaultManga(config: MangaConfig): void;
  setNotifications(value: boolean): void;
}

export const useSettings = create<SettingsState>()(
  persist(
    immer((set) => ({
      ereaderKey: defaultEReaderKey,
      toCloudDefault: false,
      defaultManga: defaultMangaConfig,
      notifications: true,
      setEReader: (key) => set({ ereaderKey: key }),
      setToCloudDefault: (value) => set({ toCloudDefault: value }),
      setLanguage: async (language) => {
        await i18n.changeLanguage(language);
        set({ language });
      },
      setDefaultManga: (config) => set({ defaultManga: config }),
      setNotifications: (value) => set({ notifications: value }),
    })),
    {
      name: 'settings',
      storage: asyncStorage,
      version: 1,
      onRehydrateStorage: () => (state) => {
        if (state?.language) void i18n.changeLanguage(state.language);
      },
    }
  )
);
