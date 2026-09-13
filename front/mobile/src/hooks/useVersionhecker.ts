import { create } from 'zustand';
import { APP_VERSION, BACKEND_API_URL, BACKEND_URL } from '../constants';
import { Platform } from 'react-native';
import { File, Paths } from 'expo-file-system';
import { startActivityAsync } from 'expo-intent-launcher';

interface State {
  showDialog: boolean;
  newVersion?: string;
  isDownloading: boolean;
  init(): Promise<void>;
  hideDialog(): void;
  installNewVersion(): Promise<void>;
}

export const useVersionChecker = create<State>((set) => ({
  showDialog: false,
  isDownloading: false,

  async init() {
    try {
      if (!APP_VERSION) return;

      const resp = await fetch(`${BACKEND_API_URL}/app/version`);
      if (!resp.ok) return;

      const version: string = await resp.json();

      set({ showDialog: version !== APP_VERSION, newVersion: version });
    } catch (e) {
      console.error('app version', e);
    }
  },

  hideDialog() {
    set({ showDialog: false });
  },

  async installNewVersion() {
    if (Platform.OS !== 'android') return;

    set({ isDownloading: true });

    try {
      const destFile = new File(`${Paths.cache}update.apk`);
      const file = await File.downloadFileAsync(`${BACKEND_URL}/app/download`, destFile);
      if (!file || !file.contentUri) {
        set({ isDownloading: false });
        return;
      }

      await startActivityAsync('android.intent.action.VIEW', {
        data: file.contentUri,
        flags: 1,
        type: 'application/vnd.android.package-archive',
      });

      set({ isDownloading: false });
    } catch (e) {
      console.error('Error installing update', e);
      set({ isDownloading: false });
    }
  },
}));
