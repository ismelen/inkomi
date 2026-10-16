import { Platform } from 'react-native';
import { File, Paths } from 'expo-file-system';
import { createDownloadResumable, getContentUriAsync } from 'expo-file-system/legacy';
import { startActivityAsync } from 'expo-intent-launcher';
import { appApi } from '../api/app-api';
import { ReleaseResponse } from '../api/types/app';
import { appVersion, backendUrl } from '../config';
import { isNewerVersion } from '../models/version';

const apkMime = 'application/vnd.android.package-archive';
const grantReadUriPermission = 1;

/** Latest release if it is newer than the installed app. Only Android installs apks. */
export async function findUpdate(): Promise<ReleaseResponse | null> {
  if (Platform.OS !== 'android') return null;
  const latest = await appApi.latest();
  return isNewerVersion(latest.version, appVersion) ? latest : null;
}

/**
 * Downloads the apk through the release-service (a redirect to GitHub Releases) and opens the
 * system installer. Android asks the user to allow "install unknown apps" the first time and
 * rejects the apk if it isn't signed with the same key.
 */
export async function installUpdate(
  release: ReleaseResponse,
  onProgress: (fraction: number) => void
): Promise<void> {
  const target = new File(Paths.cache, `inkomi-${release.version}.apk`);
  if (target.exists) target.delete();

  const download = createDownloadResumable(
    backendUrl + release.downloadPath,
    target.uri,
    {},
    ({ totalBytesWritten, totalBytesExpectedToWrite }) => {
      const total = totalBytesExpectedToWrite > 0 ? totalBytesExpectedToWrite : release.size;
      if (total > 0) onProgress(Math.min(1, totalBytesWritten / total));
    }
  );
  const result = await download.downloadAsync();
  if (!result || result.status >= 300) {
    throw new Error(`apk download failed: ${result?.status ?? 'cancelled'}`);
  }
  if (release.size > 0 && new File(result.uri).size !== release.size) {
    throw new Error('apk download incomplete');
  }

  await startActivityAsync('android.intent.action.VIEW', {
    data: await getContentUriAsync(result.uri),
    flags: grantReadUriPermission,
    type: apkMime,
  });
}
