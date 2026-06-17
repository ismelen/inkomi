import { Source } from './source';

export interface Upload {
  sources: Source[];
  uploadSettings?: UploadSettings;
}

export class UploadSettings {
  constructor(
    public mergeFiles: boolean,
    public deleteFilesAfterUpload: boolean,
    public autoUpload: boolean,
    public cloud: boolean,
    public title?: string,
    public author?: string,
    public initialVolume?: number
  ) {}

  static default(name: string, author: string): UploadSettings {
    return {
      mergeFiles: true,
      deleteFilesAfterUpload: true,
      title: name,
      author: author,
      cloud: false,
      initialVolume: 1,
      autoUpload: false,
    };
  }
}
