import { eReaderProfiles } from '../constants';
import { StorageService } from '../services/storage-service';

const EREADER_MODEL = 'eReader_model_key';
const DEFAULT_MODEL = 'KoCC';

export class eReaderModel {
  private static _inst?: eReaderModel;

  private constructor(
    private model: string,
    public name: string
  ) {}

  public static async instance(): Promise<eReaderModel> {
    if (this._inst) return this._inst;

    const data = await StorageService.GetAsync<eReaderModel>(EREADER_MODEL);
    if (!data) {
      this._inst = new eReaderModel(DEFAULT_MODEL, eReaderProfiles[DEFAULT_MODEL]);
    } else {
      this._inst = new eReaderModel(data.model, data.name);
    }

    StorageService.SetAsync(EREADER_MODEL, this._inst);
    return this._inst;
  }

  public getModel(): string {
    return this.model;
  }

  public setModel(model: string) {
    this.model = model;
    this.name = eReaderProfiles[model];

    StorageService.SetAsync(EREADER_MODEL, this);
  }
}
