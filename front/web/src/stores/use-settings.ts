import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import i18n, { SupportedLanguage } from '../i18n/i18n';
import { defaultEReaderKey } from '../models/ereaders';
import { defaultMangaConfig, MangaConfig } from '../models/manga-config';
import { localJsonStorage } from './persist-storage';

export type Theme = 'system' | 'light' | 'dark';

interface SettingsState {
  ereaderKey: string;
  /** Initial value of the Dropbox switch in the list. */
  toCloudDefault: boolean;
  language?: SupportedLanguage;
  theme: Theme;
  defaultManga: MangaConfig;
  notifications: boolean;
  setEReader(key: string): void;
  setToCloudDefault(value: boolean): void;
  setLanguage(language: SupportedLanguage): Promise<void>;
  setTheme(theme: Theme): void;
  setDefaultManga(config: MangaConfig): void;
  setNotifications(value: boolean): void;
}

function applyTheme(theme: Theme) {
  const root = document.documentElement;
  if (theme === 'system') delete root.dataset.theme;
  else root.dataset.theme = theme;
}

export const useSettings = create<SettingsState>()(
  persist(
    immer((set) => ({
      ereaderKey: defaultEReaderKey,
      toCloudDefault: false,
      theme: 'system',
      defaultManga: defaultMangaConfig,
      notifications: false,
      setEReader: (key) => set({ ereaderKey: key }),
      setToCloudDefault: (value) => set({ toCloudDefault: value }),
      setLanguage: async (language) => {
        await i18n.changeLanguage(language);
        set({ language });
      },
      setTheme: (theme) => {
        applyTheme(theme);
        set({ theme });
      },
      setDefaultManga: (config) => set({ defaultManga: config }),
      setNotifications: (value) => set({ notifications: value }),
    })),
    {
      name: 'settings',
      storage: localJsonStorage,
      version: 1,
      onRehydrateStorage: () => (state) => {
        if (state?.language) void i18n.changeLanguage(state.language);
        if (state) applyTheme(state.theme);
      },
    }
  )
);
