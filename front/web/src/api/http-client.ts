import { backendUrl } from '../config';
import { clearTokens, getTokens, reloadTokens, saveTokens } from '../services/session-storage';
import { ApiError, NetworkError } from './api-error';
import { TokensResponse } from './types/auth';

type Method = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
type Query = Record<string, string | number | undefined>;

interface RequestOptions {
  body?: unknown;
  query?: Query;
  /** Sends the bearer token and refreshes it on 401. Default true. */
  auth?: boolean;
  signal?: AbortSignal;
}

// Refresh this long before the access token expires.
const expiryMarginMs = 30_000;

let refreshing: Promise<string | null> | null = null;

function buildUrl(path: string, query?: Query): string {
  const params = Object.entries(query ?? {})
    .filter(([, value]) => value !== undefined && value !== '')
    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`);
  return backendUrl + path + (params.length ? `?${params.join('&')}` : '');
}

async function send(method: Method, path: string, opts: RequestOptions, token?: string) {
  const headers: Record<string, string> = { Accept: 'application/json' };
  if (opts.body !== undefined) headers['Content-Type'] = 'application/json';
  if (token) headers.Authorization = `Bearer ${token}`;

  try {
    return await fetch(buildUrl(path, opts.query), {
      method,
      headers,
      body: opts.body === undefined ? undefined : JSON.stringify(opts.body),
      signal: opts.signal,
    });
  } catch (err) {
    if (opts.signal?.aborted) throw err;
    throw new NetworkError(err);
  }
}

// Errors are {error: string} JSON, except the JWT middleware that answers 401 in plain text.
async function parse<T>(res: Response): Promise<T> {
  const text = await res.text();
  const isJson = res.headers.get('content-type')?.includes('application/json');
  let data: unknown = undefined;
  if (text && isJson) {
    try {
      data = JSON.parse(text);
    } catch {
      data = undefined;
    }
  }

  if (!res.ok) {
    const message =
      (data as { error?: string } | undefined)?.error ?? (text.trim() || res.statusText);
    throw new ApiError(res.status, message);
  }

  return (res.status === 204 ? undefined : data) as T;
}

async function doRefresh(): Promise<string | null> {
  const used = getTokens();
  if (!used) return null;

  // Inside the lock another tab may have rotated the refresh token already: reuse its result.
  const tokens = reloadTokens();
  if (!tokens) return null;
  if (tokens.refreshToken !== used.refreshToken) return tokens.accessToken;

  const res = await send('POST', '/auth/refresh', {
    body: { refreshToken: tokens.refreshToken },
  });
  if (res.status === 401) {
    // The refresh token is single use: a rejected one means the session is over.
    clearTokens();
    return null;
  }

  const next = await parse<TokensResponse>(res);
  saveTokens(next);
  return next.accessToken;
}

/** Only one tab of the browser refreshes at a time: the refresh token is single use. */
async function lockedRefresh(): Promise<string | null> {
  if (!navigator.locks) return doRefresh();
  return await navigator.locks.request('inkomi-refresh', doRefresh);
}

/** Refreshes the session once even if several requests (or tabs) ask at the same time. */
export function refreshAccessToken(): Promise<string | null> {
  refreshing ??= lockedRefresh().finally(() => {
    refreshing = null;
  });
  return refreshing;
}

/** Access token valid for at least the next seconds, refreshing it if needed. */
export async function getValidAccessToken(): Promise<string | null> {
  const tokens = getTokens();
  if (!tokens) return null;

  const expiresAt = Date.parse(tokens.accessTokenExpiresAt);
  if (Number.isNaN(expiresAt) || expiresAt - expiryMarginMs > Date.now()) {
    return tokens.accessToken;
  }
  return refreshAccessToken();
}

export async function request<T>(
  method: Method,
  path: string,
  opts: RequestOptions = {}
): Promise<T> {
  const auth = opts.auth ?? true;
  if (!auth) return parse<T>(await send(method, path, opts));

  const token = await getValidAccessToken();
  if (!token) throw new ApiError(401, 'invalid or expired token');

  const res = await send(method, path, opts, token);
  if (res.status !== 401) return parse<T>(res);

  const refreshed = await refreshAccessToken();
  if (!refreshed) throw new ApiError(401, 'invalid or expired token');
  return parse<T>(await send(method, path, opts, refreshed));
}

export const http = {
  get: <T>(path: string, opts?: RequestOptions) => request<T>('GET', path, opts),
  post: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>('POST', path, { ...opts, body }),
  put: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>('PUT', path, { ...opts, body }),
  patch: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>('PATCH', path, { ...opts, body }),
  delete: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>('DELETE', path, { ...opts, body }),
};
