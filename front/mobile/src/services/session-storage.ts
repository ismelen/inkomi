import * as SecureStore from 'expo-secure-store';
import { randomUUID } from 'expo-crypto';
import { TokensResponse } from '../api/types/auth';

const tokensKey = 'session_tokens';
const deviceIdKey = 'device_id';

export type Tokens = TokensResponse;
type Listener = (tokens: Tokens | null) => void;

let current: Tokens | null = null;
let loaded = false;
const listeners = new Set<Listener>();

function emit() {
  listeners.forEach((listener) => listener(current));
}

/** Loads the tokens kept in SecureStore; later reads are synchronous. */
export async function loadTokens(): Promise<Tokens | null> {
  if (loaded) return current;

  const raw = await SecureStore.getItemAsync(tokensKey);
  try {
    current = raw ? (JSON.parse(raw) as Tokens) : null;
  } catch {
    current = null;
  }
  loaded = true;
  return current;
}

export function getTokens(): Tokens | null {
  return current;
}

export async function saveTokens(tokens: Tokens): Promise<void> {
  current = tokens;
  loaded = true;
  await SecureStore.setItemAsync(tokensKey, JSON.stringify(tokens));
  emit();
}

export async function clearTokens(): Promise<void> {
  current = null;
  loaded = true;
  await SecureStore.deleteItemAsync(tokensKey);
  emit();
}

export function onTokensChange(listener: Listener): () => void {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

export async function getDeviceId(): Promise<string> {
  const stored = await SecureStore.getItemAsync(deviceIdKey);
  if (stored) return stored;

  const id = randomUUID();
  await SecureStore.setItemAsync(deviceIdKey, id);
  return id;
}
