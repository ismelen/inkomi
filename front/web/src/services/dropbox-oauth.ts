import { dropboxApiKey } from '../config';

/** What to do with the refresh token when Dropbox redirects back. */
export type DropboxIntent = 'login' | 'link';

interface PendingAuth {
  verifier: string;
  state: string;
  intent: DropboxIntent;
  returnTo: string;
}

const pendingKey = 'dropbox_oauth';
export const dropboxCallbackPath = '/oauth/dropbox';

function redirectUri(): string {
  return window.location.origin + dropboxCallbackPath;
}

function base64Url(bytes: Uint8Array): string {
  return btoa(String.fromCharCode(...bytes))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

async function generatePKCE() {
  const verifier = base64Url(crypto.getRandomValues(new Uint8Array(32)));
  const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(verifier));
  return { verifier, challenge: base64Url(new Uint8Array(digest)) };
}

/**
 * Starts the Dropbox PKCE flow with a full page redirect. The callback page finishes it with
 * completeDropboxAuth; the auth-service keeps the offline refresh token.
 */
export async function startDropboxAuth(intent: DropboxIntent): Promise<void> {
  const { verifier, challenge } = await generatePKCE();
  const state = base64Url(crypto.getRandomValues(new Uint8Array(16)));
  const pending: PendingAuth = {
    verifier,
    state,
    intent,
    returnTo: window.location.pathname + window.location.search,
  };
  sessionStorage.setItem(pendingKey, JSON.stringify(pending));

  const params = new URLSearchParams({
    client_id: dropboxApiKey,
    redirect_uri: redirectUri(),
    response_type: 'code',
    code_challenge: challenge,
    code_challenge_method: 'S256',
    token_access_type: 'offline',
    scope: 'files.content.write files.metadata.read account_info.read',
    state,
  });
  window.location.assign(`https://www.dropbox.com/oauth2/authorize?${params}`);
}

export interface DropboxAuthResult {
  intent: DropboxIntent;
  returnTo: string;
  /** Undefined when the user cancelled on Dropbox. */
  refreshToken?: string;
}

/** Exchanges the code of the callback url. Null when there was no flow in progress. */
export async function completeDropboxAuth(search: string): Promise<DropboxAuthResult | null> {
  const raw = sessionStorage.getItem(pendingKey);
  sessionStorage.removeItem(pendingKey);
  if (!raw) return null;

  const pending = JSON.parse(raw) as PendingAuth;
  const params = new URLSearchParams(search);
  const code = params.get('code');
  const result = { intent: pending.intent, returnTo: pending.returnTo };
  if (!code || params.get('state') !== pending.state) return result;

  const res = await fetch('https://api.dropboxapi.com/oauth2/token', {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      code,
      grant_type: 'authorization_code',
      client_id: dropboxApiKey,
      redirect_uri: redirectUri(),
      code_verifier: pending.verifier,
    }).toString(),
  });
  if (!res.ok) throw new Error(`dropbox token exchange: ${res.status}`);

  const { refresh_token } = (await res.json()) as { refresh_token?: string };
  return { ...result, refreshToken: refresh_token };
}
