import { GoogleSignin } from '@react-native-google-signin/google-signin';
import { useGoogleDrivePicker } from '../components/modals/google-drive-picker-modal';
import { WEB_CLIENT_ID, WEB_CLIENT_SECRET } from '../constants';
import { StorageService } from '../services/storage-service';

const CLOUD = 'cloud_key';

interface TokenData {
  token?: string;
  refresh?: string;
  expirationDate?: Date;
}

export class Cloud {
  private static _instance?: Cloud;

  private constructor(
    public email?: string,
    public folderName?: string,
    private token?: TokenData,
    private folderId?: string
  ) {}

  static async instance(): Promise<Cloud> {
    if (this._instance) return this._instance;
    const json = await StorageService.GetSecureAsync(CLOUD);

    if (!json) {
      this._instance = new Cloud();
      return this._instance;
    }

    const data: Cloud = JSON.parse(json);
    this._instance = new Cloud(
      data.email,
      data.folderName,
      {
        ...data.token,
        expirationDate: data.token?.expirationDate
          ? new Date(data.token.expirationDate)
          : undefined,
      },
      data.folderId
    );
    return this._instance!;
  }

  public async setAccount() {
    await this.requestToken();
    StorageService.SetSecureAsync(CLOUD, JSON.stringify(this));
  }

  public async setFolder() {
    const json = JSON.stringify(this);
    console.log(json);
    if (!this.token?.token) {
      alert('No account specified');
      return;
    }

    const values = await this.requestFolder();
    this.folderName = values?.name;
    this.folderId = values?.id;
    StorageService.SetSecureAsync(CLOUD, JSON.stringify(this));
  }

  public async check(): Promise<boolean> {
    const tokenUpdated = await this.updateToken();
    if (!tokenUpdated) return false;

    const folderUpdated = await this.updateFolderId();
    if (!folderUpdated) return false;

    StorageService.SetSecureAsync(CLOUD, JSON.stringify(this));
    return true;
  }

  getToken(): string {
    return this.token?.token!;
  }

  getFolderId(): string {
    return this.folderId!;
  }

  private async updateFolderId(): Promise<boolean> {
    if (this.folderId) return true;
    if (!this.token?.token) return false;

    const values = await this.requestFolder();

    return values !== undefined;
  }

  private async requestFolder(): Promise<
    | {
        id: string;
        name: string;
      }
    | undefined
  > {
    const values = await useGoogleDrivePicker.getState().show(this.token?.token!);
    this.folderId = values?.id;
    this.folderName = values?.name;

    return values;
  }

  private async updateToken(): Promise<boolean> {
    const now = new Date();

    if (this.token?.expirationDate) {
      if (this.token.expirationDate.getTime() > now.getTime()) {
        return true;
      }

      this.token = await this.refreshToken();
      return this.token !== undefined;
    }

    return await this.requestToken();
  }

  private async refreshToken(): Promise<TokenData | undefined> {
    if (!this.token?.refresh) return;

    const response = await fetch('https://oauth2.googleapis.com/token', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: new URLSearchParams({
        client_id: WEB_CLIENT_ID,
        refresh_token: this.token.refresh,
        grant_type: 'refresh_token',
      }).toString(),
    });

    const tokens = await response.json();
    console.log(tokens);
    if (tokens.error) return;

    const expiration = new Date();
    expiration.setHours(expiration.getHours() + 1);

    return {
      token: tokens.access_token,
      refresh: tokens.refresh_token,
      expirationDate: expiration,
    };
  }

  private async requestToken(): Promise<boolean> {
    GoogleSignin.configure({
      webClientId: WEB_CLIENT_ID,
      offlineAccess: true,
      scopes: [
        'https://www.googleapis.com/auth/drive',
        'https://www.googleapis.com/auth/userinfo.profile',
        'https://www.googleapis.com/auth/userinfo.email',
        'openid',
      ],
      forceCodeForRefreshToken: true,
    });

    await GoogleSignin.hasPlayServices();
    const result = await GoogleSignin.signIn();
    if (result.type !== 'success') return false;

    const { data } = result;
    if (data.serverAuthCode === null) return false;

    const response = await fetch('https://oauth2.googleapis.com/token', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: new URLSearchParams({
        code: data.serverAuthCode,
        client_id: WEB_CLIENT_ID,
        client_secret: WEB_CLIENT_SECRET,
        grant_type: 'authorization_code',
        redirect_uri: '',
      }).toString(),
    });

    const tokens = await response.json();
    if (tokens.error) return false;

    const expiration = new Date();
    expiration.setHours(expiration.getHours() + 1);

    this.email = data.user.email;
    this.token = {
      token: tokens.access_token,
      refresh: tokens.refresh_token,
      expirationDate: expiration,
    };

    return true;
  }
}
