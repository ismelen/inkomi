import { pickDirectory } from '@react-native-documents/picker';
import { create } from 'zustand';
import { MonitoredFolder } from '../models/monitored-folder';
import { UploadSettings } from '../models/upload';
import { FilesystemService } from '../services/filesystem-service';
import { KepubifyService } from '../services/kepubify-service';
import { MangaConvertService } from '../services/manga-convert-service';
import { StorageService } from '../services/storage-service';
import { useQueue } from './use-queue';

interface State {
  folders: MonitoredFolder[];
  loading: boolean;
  fetchMonitoredFolders(): Promise<void>;
  addFolder(): Promise<void>;
  deleteFolder(idx: number): Promise<void>;
  updateFolderSettings(settings: UploadSettings, kepubify: boolean, idx: number): Promise<void>;
  updateFolder(folder: MonitoredFolder, idx: number): Promise<void>;
  autoUpload(): Promise<void>;
}

const FOLDERS = 'folders_key';

export const useMonitoredFolders = create<State>((set, get) => ({
  folders: [],
  loading: false,

  async fetchMonitoredFolders() {
    const monitoredFolders = await StorageService.GetAsync<MonitoredFolder[]>(FOLDERS);
    if (!monitoredFolders) return;

    // TODO: Check, add lastUpdatedChildrens ??
    for (let folder of monitoredFolders) {
      const source = await FilesystemService.getMonitorizedFolder(folder.source.path);
      if (!folder.uploaded || (folder.source.children?.length ?? 0) === 0) {
        folder.source = source;
        folder.uploaded = false;
        continue;
      }

      const lastUpdatePaths = new Set(folder.lastUploadedPahts);
      folder.source.children = source.children?.filter((path) => !lastUpdatePaths.has(path));
    }

    StorageService.SetAsync(FOLDERS, monitoredFolders);

    set({ folders: monitoredFolders });

    get().autoUpload();
  },

  async addFolder() {
    const dir = await pickDirectory({
      requestLongTermAccess: true,
    });
    const folder = await FilesystemService.getMonitorizedFolder(dir.uri);
    const settings = UploadSettings.default(folder.name, '');
    const newFolders = [
      ...get().folders,
      {
        source: folder,
        settings: settings,
        uploaded: false,
        kepubify: false,
        lastUploadedPahts: [],
      } as MonitoredFolder,
    ];
    set({ folders: newFolders });
    StorageService.SetAsync(FOLDERS, newFolders);
  },

  async deleteFolder(idx: number) {
    let folders = get().folders;
    folders = folders.filter((_, i) => i !== idx);

    set({ folders: [...folders] });
    StorageService.SetAsync(FOLDERS, folders);
  },

  async updateFolderSettings(settings: UploadSettings, kepubify: boolean, idx: number) {
    const folders = get().folders;
    folders[idx].settings = settings;
    folders[idx].kepubify = kepubify;

    set({ folders: [...folders] });
    StorageService.SetAsync(FOLDERS, folders);
  },

  async updateFolder(folder: MonitoredFolder, idx: number) {
    const folders = get().folders;
    folders[idx] = folder;

    set({ folders: [...folders] });
    StorageService.SetAsync(FOLDERS, folders);
  },

  async autoUpload() {
    const folders = get().folders;
    set({ loading: true });

    for (let folder of folders) {
      if (!folder.settings.autoUpload) continue;
      if (folder.kepubify) {
        const request = await KepubifyService.convert(folder.source.children!, folder.settings);
        if (!request) continue;
        if (folder.settings.deleteFilesAfterUpload) {
          folder.source.children = [];
          folder.uploaded = true;
        }
        continue;
      }

      const request = await MangaConvertService.convert(folder.source.children!, folder.settings);
      if (!request) continue;
      useQueue.getState().add(request);

      folder.settings.initialVolume! += request?.times.length;
      if (folder.settings.deleteFilesAfterUpload) {
        folder.source.children = [];
      }
      folder.uploaded = true;
    }

    await StorageService.SetAsync(FOLDERS, folders);

    set({ loading: false, folders: [...folders] });
  },
}));



