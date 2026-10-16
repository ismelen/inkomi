import * as Linking from 'expo-linking';
import * as WebBrowser from 'expo-web-browser';
import {
  CryptoDigestAlgorithm,
  CryptoEncoding,
  digestStringAsync,
  getRandomBytes,
} from 'expo-crypto';
import { dropboxApiKey } from '../config';

WebBrowser.maybeCompleteAuthSession();

function base64Url(value: string): string {
  return value.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

async function generatePKCE() {
  const verifier = base64Url(btoa(String.fromCharCode(...getRandomBytes(32))));
  const digest = await digestStringAsync(CryptoDigestAlgorithm.SHA256, verifier, {
    encoding: CryptoEncoding.BASE64,
  });
  return { verifier, challenge: base64Url(digest) };
}

/**
 * Runs the Dropbox PKCE flow in the browser and returns an offline refresh token, which the
 * auth-service keeps (POST /auth/dropbox or PUT /users/me/dropbox). Undefined when cancelled.
 */
export async function getDropboxRefreshToken(): Promise<string | undefined> {
  const redirectUri = Linking.createURL('oauth');
  const { verifier, challenge } = await generatePKCE();

  const authUrl =
    'https://www.dropbox.com/oauth2/authorize' +
    `?client_id=${dropboxApiKey}` +
    `&redirect_uri=${encodeURIComponent(redirectUri)}` +
    '&response_type=code' +
    `&code_challenge=${challenge}` +
    '&code_challenge_method=S256' +
    '&token_access_type=offline' +
    '&scope=files.content.write%20files.metadata.read%20account_info.read';

  const result = await WebBrowser.openAuthSessionAsync(authUrl, redirectUri);
  if (result.type !== 'success') return undefined;

  const code = Linking.parse(result.url).queryParams?.code;
  if (typeof code !== 'string') return undefined;

  const res = await fetch('https://api.dropboxapi.com/oauth2/token', {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      code,
      grant_type: 'authorization_code',
      client_id: dropboxApiKey,
      redirect_uri: redirectUri,
      code_verifier: verifier,
    }).toString(),
  });
  if (!res.ok) throw new Error(`dropbox token exchange: ${res.status}`);

  const { refresh_token } = (await res.json()) as { refresh_token?: string };
  return refresh_token;
}
