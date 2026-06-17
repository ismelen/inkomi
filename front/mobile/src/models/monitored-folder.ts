import { Source } from './source';
import { UploadSettings } from './upload';

export interface MonitoredFolder {
  source: Source;
  settings: UploadSettings;
  uploaded: boolean;
  kepubify: boolean;
  lastUploadedPahts: string[];
}
