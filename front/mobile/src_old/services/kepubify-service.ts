import mime from 'mime';
import { copyToCache } from '../../modules/file-handler';
import { Cloud } from '../models/cloud';
import { QueueRequest, QueueTime } from '../models/queue';
import { UploadSettings } from '../models/upload';
import { DownloadService } from './download-service';
import { FilesystemService } from './filesystem-service';
import { BACKENDD_URL } from '../constants';

export class KepubifyService {
  public static async convert(
    paths: string[],
    settings: UploadSettings
  ): Promise<QueueRequest | undefined> {
    try {
      const form = new FormData();

      for (let path of paths) {
        const name = decodeURIComponent(path).split('/').pop();
        if (!name) {
          alert(`Error with file: ${path}`);
          return;
        }
        const cachePath = await copyToCache(path, name);

        form.append('files', {
          name: name,
          type: mime.getType(cachePath),
          uri: cachePath,
        } as unknown as Blob);
      }

      if (settings.cloud) {
        const cloud = await Cloud.instance();
        if (!cloud.check()) {
          alert('No cloud data, converting in local');
        } else {
          form.append('cloudFolder', cloud.getFolderId());
          form.append('cloudToken', cloud.getToken());
        }
      }

      const response = await fetch(`${BACKENDD_URL}/kepubify`, {
        body: form,
        method: 'POST',
      });

      const json = await response.json();
      console.log(json);
      if (!response.ok) {
        alert(json.error);
        return;
      }

      if (settings.deleteFilesAfterUpload && response.ok) {
        for (let path of paths) {
          FilesystemService.deleteFile(path);
        }
      }

      if (!settings.cloud) {
        for (let path of json.paths as string[]) {
          DownloadService.download(path);
        }
      }

      return {
        sources: [],
        settings: settings,
        times: json.paths.map(
          (e: any) =>
            ({
              path: e.path,
              endTime: new Date(e.endTime),
            }) as QueueTime
        ),
      };
    } catch (e) {
      const msg = (e as Error).message;
      alert(msg);
      console.log(msg);
    }
  }
}
