import { Platform } from 'react-native';
import { Directory, File, Paths } from 'expo-file-system';
import * as Sharing from 'expo-sharing';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { uploadsApi } from '../api/uploads-api';

const downloadDirKey = 'download_dir_uri';

const mimeTypes: Record<string, string> = {
  epub: 'application/epub+zip',
  pdf: 'application/pdf',
  cbz: 'application/vnd.comicbook+zip',
};

function mimeOf(filename: string): string {
  return mimeTypes[filename.split('.').pop()?.toLowerCase() ?? ''] ?? 'application/octet-stream';
}

/**
 * Asks for a fresh presigned link (never cached: it expires after 15 min) and downloads the file
 * into the cache. Errors (409 not ready, 404 purged) are thrown to be shown, never retried.
 */
async function downloadToCache(sourceId: string): Promise<File> {
  const link = await uploadsApi.download(sourceId);
  const dir = new Directory(Paths.cache, 'downloads', sourceId);
  dir.create({ idempotent: true, intermediates: true });

  const target = new File(dir, link.filename);
  if (target.exists) target.delete();
  const downloaded = await File.downloadFileAsync(link.url, target);
  return new File(downloaded.uri);
}

export async function shareSource(sourceId: string): Promise<void> {
  const file = await downloadToCache(sourceId);
  await Sharing.shareAsync(file.uri, { mimeType: mimeOf(file.name), dialogTitle: file.name });
}

async function pickDownloadDir(): Promise<Directory> {
  const picked = await Directory.pickDirectoryAsync();
  await AsyncStorage.setItem(downloadDirKey, picked.uri);
  return new Directory(picked.uri);
}

async function writeTo(dir: Directory, file: File) {
  const target = dir.createFile(file.name, mimeOf(file.name));
  target.write(await file.bytes());
}

/**
 * Android: saves into a folder picked once (SAF) and remembered. iOS: the share sheet, where the
 * user picks "Save to Files". Returns false when the user cancels the folder picker.
 */
export async function saveSource(sourceId: string): Promise<boolean> {
  const file = await downloadToCache(sourceId);
  if (Platform.OS !== 'android') {
    await Sharing.shareAsync(file.uri, { mimeType: mimeOf(file.name), UTI: 'public.data' });
    return true;
  }

  const saved = await AsyncStorage.getItem(downloadDirKey);
  if (saved) {
    try {
      await writeTo(new Directory(saved), file);
      return true;
    } catch {
      // The permission on the remembered folder was lost: ask again.
      await AsyncStorage.removeItem(downloadDirKey);
    }
  }

  let dir: Directory;
  try {
    dir = await pickDownloadDir();
  } catch {
    return false;
  }
  await writeTo(dir, file);
  return true;
}
