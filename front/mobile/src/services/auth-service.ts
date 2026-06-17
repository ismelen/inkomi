import { GoogleSignin } from '@react-native-google-signin/google-signin';
import Constants from 'expo-constants';
import { StorageService } from './storage-service';
import { WEB_CLIENT_ID, WEB_CLIENT_SECRET } from '../constants';

const TOKEN_KEY = 'token';


interface TokenData {
  token?: string;
  refresh?: string;
  expirationDate?: Date;
}

export class AuthService {
  private data?: TokenData;

  public constructor() {
    this.init();
  }

  private async init() {
    const json = await StorageService.GetSecureAsync(TOKEN_KEY);
    if (!json) return;

    this.data = JSON.parse(json) as TokenData;
  }

  public async getToken(): Promise<string | undefined> {
    const now = new Date();
    if (this.data?.expirationDate) {
      if (new Date(this.data.expirationDate).getTime() > now.getTime()) {
        return this.data.token;
      }
      this.data = await this.refreshToken();
      StorageService.SetAsync(TOKEN_KEY, JSON.stringify(this.data));

      return this.data?.token;
    }

    this.data = await this.requestToken();
    await StorageService.SetSecureAsync(TOKEN_KEY, JSON.stringify(this.data));

    return this.data?.token;
  }

  private async requestToken(): Promise<TokenData | undefined> {
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
    if (result.type !== 'success') return;

    const now = new Date();
    now.setHours(now.getHours() + 1);

    const { data } = result;
    if (data.serverAuthCode === null) return;

    // Debes intercambiar el código por tokens
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
    if (tokens.error) return;

    return {
      token: tokens.access_token,
      refresh: tokens.refresh_token,
      expirationDate: now,
    };
  }

  private async refreshToken(): Promise<TokenData | undefined> {
    if (!this.data?.refresh) return;

    const response = await fetch('https://oauth2.googleapis.com/token', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: new URLSearchParams({
        client_id: WEB_CLIENT_ID,
        refresh_token: this.data.refresh,
        grant_type: 'refresh_token',
      }).toString(),
    });

    const tokens = await response.json();

    if (tokens.error) return;

    const now = new Date();
    now.setHours(now.getHours() + 1);

    return {
      token: tokens.access_token,
      refresh: tokens.refresh_token,
      expirationDate: now,
    };
  }
}
