import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import i18n, { SupportedLanguage } from '../i18n/i18n';
import { defaultEReaderKey } from '../models/ereaders';
import { defaultMangaConfig, MangaConfig } from '../models/manga-config';
import { MangaProfile, newProfileId } from '../models/manga-profile';
import { asyncStorage } from './persist-storage';

interface SettingsState {
  ereaderKey: string;
  /** Initial value of the Dropbox switch in Home. */
  toCloudDefault: boolean;
  language?: SupportedLanguage;
  defaultManga: MangaConfig;
  notifications: boolean;
  /** Manga configs saved with a name, to reuse in any item, group or the default. */
  mangaProfiles: MangaProfile[];
  setEReader(key: string): void;
  setToCloudDefault(value: boolean): void;
  setLanguage(language: SupportedLanguage): Promise<void>;
  setDefaultManga(config: MangaConfig): void;
  setNotifications(value: boolean): void;
  /** Saves config as name, replacing the profile with that name if there is one. */
  saveMangaProfile(name: string, config: MangaConfig): void;
  renameMangaProfile(id: string, name: string): void;
  deleteMangaProfile(id: string): void;
}

export const useSettings = create<SettingsState>()(
  persist(
    immer((set) => ({
      ereaderKey: defaultEReaderKey,
      toCloudDefault: false,
      defaultManga: defaultMangaConfig,
      notifications: true,
      mangaProfiles: [],
      setEReader: (key) => set({ ereaderKey: key }),
      setToCloudDefault: (value) => set({ toCloudDefault: value }),
      setLanguage: async (language) => {
        await i18n.changeLanguage(language);
        set({ language });
      },
      setDefaultManga: (config) => set({ defaultManga: config }),
      setNotifications: (value) => set({ notifications: value }),
      saveMangaProfile: (name, config) =>
        set((s) => {
          const existing = s.mangaProfiles.find((p) => p.name === name);
          if (existing) existing.config = config;
          else s.mangaProfiles.push({ id: newProfileId(), name, config });
        }),
      renameMangaProfile: (id, name) =>
        set((s) => {
          const profile = s.mangaProfiles.find((p) => p.id === id);
          if (profile) profile.name = name;
        }),
      deleteMangaProfile: (id) =>
        set((s) => {
          s.mangaProfiles = s.mangaProfiles.filter((p) => p.id !== id);
        }),
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
