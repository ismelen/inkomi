import { TokensResponse } from '../api/types/auth';

const tokensKey = 'session_tokens';
const deviceIdKey = 'device_id';

export type Tokens = TokensResponse;
type Listener = (tokens: Tokens | null) => void;

const listeners = new Set<Listener>();

function read(): Tokens | null {
  try {
    const raw = localStorage.getItem(tokensKey);
    return raw ? (JSON.parse(raw) as Tokens) : null;
  } catch {
    return null;
  }
}

let current: Tokens | null = read();

function emit() {
  listeners.forEach((listener) => listener(current));
}

// Every tab shares the session: a refresh or a sign out in another tab applies here too.
window.addEventListener('storage', (event) => {
  if (event.key !== tokensKey && event.key !== null) return;
  current = read();
  emit();
});

export function getTokens(): Tokens | null {
  return current;
}

/** Reads the tokens again from storage: another tab may have rotated them. */
export function reloadTokens(): Tokens | null {
  current = read();
  return current;
}

export function saveTokens(tokens: Tokens): void {
  current = tokens;
  localStorage.setItem(tokensKey, JSON.stringify(tokens));
  emit();
}

export function clearTokens(): void {
  current = null;
  localStorage.removeItem(tokensKey);
  emit();
}

export function onTokensChange(listener: Listener): () => void {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

export function getDeviceId(): string {
  const stored = localStorage.getItem(deviceIdKey);
  if (stored) return stored;

  const id = crypto.randomUUID();
  localStorage.setItem(deviceIdKey, id);
  return id;
}
