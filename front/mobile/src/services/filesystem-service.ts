import { pickDirectory } from '@react-native-documents/picker';
import { getDocumentAsync } from 'expo-document-picker';
import * as FS from 'expo-file-system/legacy';
import { Source } from '../models/source';

export class FilesystemService {
  static async getMonitorizedFolder(dir: string): Promise<Source> {
    const files = await FS.StorageAccessFramework.readDirectoryAsync(dir);

    return {
      name: decodeURIComponent(dir).split('/').pop() ?? 'ERROR',
      path: dir,
      children: files,
    };
  }

  static async pickFolder(): Promise<Source | undefined> {
    const dir = await pickDirectory({
      requestLongTermAccess: true,
    });
    if (dir.bookmarkStatus !== 'success') return;

    return this.getMonitorizedFolder(dir.uri);
  }

  static async pickFiles(): Promise<Source[]> {
    const result = await getDocumentAsync({
      copyToCacheDirectory: false,
      multiple: true,
    });
    if (result.canceled) return [];

    const srcs: Source[] = [];
    for (var file of result.assets) {
      srcs.push({
        name: file.name,
        path: file.uri,
        // size: file.size
      });
    }

    return srcs;
  }

  static async deleteFile(path: string) {
    await FS.deleteAsync(path, { idempotent: true });
  }

  // static async readDirectory(folder: Source): Promise<Source> {
  //   const files = await FS.StorageAccessFramework.readDirectoryAsync(folder.path);

  //   folder.lastSync = new Date();
  //   folder.status = 'Pending';
  //   folder.synchronized = true;
  //   folder.pendingFilesAmount = files.length - folder.lastFilesAmount;

  //   return folder;
  // }

  // static async getDirectoryData(dir: string): Promise<Source> {
  //   const files = await FS.StorageAccessFramework.readDirectoryAsync(dir);

  //   const result: Source = {
  //     name: decodeURIComponent(dir).split('/').pop() ?? 'Error',
  //     path: dir,
  //     status: 'Pending',
  //     synchronized: true,
  //     watching: false,
  //     lastSync: new Date(),
  //     pendingFilesAmount: files?.length ?? 0,
  //     lastFilesAmount: 0,
  //   };

  //   return result;
  // }

  // static formatBytes(bytes: number): string {
  //   if (bytes === 0) return '0 Bytes';

  //   const k = 1024; // O usa 1000 si prefieres el sistema decimal
  //   const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB'];

  //   // Calculamos a qué índice de "sizes" corresponde
  //   const i = Math.floor(Math.log(bytes) / Math.log(k));

  //   // Convertimos el valor y concatenamos la unidad
  //   return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
  // }
}
