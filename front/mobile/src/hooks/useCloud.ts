import { create } from 'zustand';
import { StorageService } from '../services/storage-service';
import { BACKENDD_URL, DROPBOX_API_KEY, WEB_CLIENT_ID, WEB_CLIENT_SECRET } from '../constants';
import * as Linking from 'expo-linking';
import * as WebBrowser from 'expo-web-browser';
import {
  CryptoDigestAlgorithm,
  CryptoEncoding,
  digestStringAsync,
  getRandomBytes,
} from 'expo-crypto';

const REFRESH_KEY = 'oauth_refresh_token';
const TOKEN_KEY = 'oauth_token';
const EXPIRES_AT_KEY = 'oauth_expires_at_token';
const FOLDER_PATH_KEY = 'folder_path_key';

interface State {
  refresh?: string;
  token?: string;
  expiresAt: number;
  folderId?: string;
  getToken(): Promise<string | undefined>;
  getFolder(): Promise<string | undefined>;
  init(): Promise<void>;
  showDialog: boolean;
  onFolderSelect?(data?: string): void;
}

export const useCloud = create<State>((set, get) => ({
  expiresAt: 0,
  showDialog: false,

  async init() {
    const refresh = await StorageService.GetSecureAsync(REFRESH_KEY);
    const token = await StorageService.GetSecureAsync(TOKEN_KEY);
    const expiresAt = await StorageService.GetSecureAsync(EXPIRES_AT_KEY);
    const folderPath = await StorageService.GetSecureAsync(FOLDER_PATH_KEY);

    set({ refresh, token, expiresAt: Number(expiresAt ?? '0'), folderId: folderPath });
  },

  async getFolder(): Promise<string | undefined> {
    const { folderId, token } = get();
    if (folderId) return folderId;
    if (!token) await get().getToken();

    const path = await new Promise<string | undefined>((resolve) => {
      set({ onFolderSelect: resolve, showDialog: true });
    });
    set({ showDialog: false, onFolderSelect: undefined });

    if (!path) return;

    set({ folderId: path });
    StorageService.SetSecureAsync(FOLDER_PATH_KEY, path ?? '');

    return path;
  },

  async getToken(): Promise<string | undefined> {
    const { refresh, token, expiresAt } = get();

    if (expiresAt <= Date.now() || !token) {
      const tokens = await (!refresh ? login() : refreshToken(refresh));
      if (!tokens) return;

      set({
        token: tokens.token,
        refresh: tokens.refresh,
        expiresAt: tokens.expiresAt,
      });

      StorageService.SetSecureAsync(TOKEN_KEY, tokens.token || '');
      StorageService.SetSecureAsync(REFRESH_KEY, tokens.refresh || '');
      StorageService.SetSecureAsync(EXPIRES_AT_KEY, String(tokens.expiresAt ?? 0));

      return tokens.token;
    }

    return token;
  },
}));

interface OAuthData {
  token?: string;
  refresh?: string;
  expiresAt?: number;
}

WebBrowser.maybeCompleteAuthSession();

async function login(): Promise<OAuthData | undefined> {
  const redirectUri = Linking.createURL('oauth');

  const { codeVerifier, codeChallenge } = await generatePKCE();

  const authUrl =
    `https://www.dropbox.com/oauth2/authorize` +
    `?client_id=${DROPBOX_API_KEY}` +
    `&redirect_uri=${encodeURIComponent(redirectUri)}` +
    `&response_type=code` +
    `&code_challenge=${codeChallenge}` +
    `&code_challenge_method=S256` +
    `&token_access_type=offline` +
    `&scope=files.content.write%20files.metadata.read`;

  const result = await WebBrowser.openAuthSessionAsync(authUrl, redirectUri);

  if (result.type !== 'success') return undefined;

  const parsed = Linking.parse(result.url);
  const code = parsed.queryParams?.code as string | undefined;
  if (!code) return undefined;

  const res = await fetch('https://api.dropboxapi.com/oauth2/token', {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      code,
      grant_type: 'authorization_code',
      client_id: DROPBOX_API_KEY,
      redirect_uri: redirectUri,
      code_verifier: codeVerifier,
    }).toString(),
  });

  if (!res.ok) return undefined;

  const { access_token, refresh_token, expires_in } = await res.json();

  return {
    token: access_token,
    refresh: refresh_token,
    expiresAt: Date.now() + expires_in * 1000,
  };
}

async function generatePKCE(): Promise<{ codeVerifier: string; codeChallenge: string }> {
  const randomBytes = getRandomBytes(32);
  const codeVerifier = btoa(String.fromCharCode(...randomBytes))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');

  const digest = await digestStringAsync(CryptoDigestAlgorithm.SHA256, codeVerifier, {
    encoding: CryptoEncoding.BASE64,
  });

  const codeChallenge = digest.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');

  return { codeVerifier, codeChallenge };
}

async function refreshToken(refresh: string): Promise<OAuthData | undefined> {
  const res = await fetch('https://api.dropboxapi.com/oauth2/token', {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'refresh_token',
      refresh_token: refresh,
      client_id: DROPBOX_API_KEY,
    }).toString(),
  });

  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error_description ?? 'Failed to refresh Dropbox token');
  }

  const { access_token, expires_in, refresh_token } = await res.json();

  return {
    expiresAt: Date.now() + expires_in * 1000,
    token: access_token,
    refresh: refresh_token,
  };
}
