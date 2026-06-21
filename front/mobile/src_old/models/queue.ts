import { Source } from './source';
import { UploadSettings } from './upload';

export interface QueueRequest {
  sources: Source[];
  settings: UploadSettings;
  times: QueueTime[];
}

export interface QueueTime {
  path: string;
  endTime?: Date;
}
