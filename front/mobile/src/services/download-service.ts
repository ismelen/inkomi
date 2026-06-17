import { File, Paths } from 'expo-file-system';
import { isAvailableAsync, shareAsync } from 'expo-sharing';
import { BACKENDD_URL } from '../constants';


export class DownloadService {
  public static async download(path: string): Promise<boolean> {
    try {
      const filename = path.split('/').pop()!;
      const dst = new File(Paths.document, filename);

      const { uri } = await File.downloadFileAsync(`${BACKENDD_URL}/download/${path}`, dst);

      console.log('downloaded at:', uri);

      if (await isAvailableAsync()) {
        await shareAsync(uri);
      }

      return true;
    } catch (e) {
      alert((e as Error).message);
      return false;
    }
  }
}
