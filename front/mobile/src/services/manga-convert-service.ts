import mime from 'mime';
import { copyToCache } from '../../modules/file-handler';
import { Cloud } from '../models/cloud';
import { eReaderModel } from '../models/e-reader-model';
import { QueueRequest, QueueTime } from '../models/queue';
import { UploadSettings } from '../models/upload';
import { FilesystemService } from './filesystem-service';
import { BACKENDD_URL } from '../constants';

export class MangaConvertService {
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

      const profile = await eReaderModel.instance();

      form.append('profile', profile.getModel());
      form.append('format', 'epub');
      form.append('author', settings.author ?? '');
      form.append('title', settings.title ?? '');
      form.append('merge', settings.mergeFiles.toString());
      form.append('firstVolumeNum', (settings.initialVolume ?? 1).toString());

      if (settings.cloud) {
        const cloud = await Cloud.instance();
        if (!cloud.check()) {
          alert('No cloud data, converting in local');
        } else {
          form.append('cloudFolder', cloud.getFolderId());
          form.append('cloudToken', cloud.getToken());
        }
      }

      const response = await fetch(`${BACKENDD_URL}/manga/convert`, {
        body: form,
        method: 'POST',
      });

      const json = await response.json();
      if (!response.ok) {
        alert(json.error);
      }

      if (settings.deleteFilesAfterUpload && response.ok) {
        for (let path of paths) {
          FilesystemService.deleteFile(path);
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
