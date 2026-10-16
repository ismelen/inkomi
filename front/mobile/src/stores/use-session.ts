import { create } from 'zustand';
import { authApi } from '../api/auth-api';
import { TokensResponse } from '../api/types/auth';
import { clearActivity } from '../services/activity-db';
import { disablePush } from '../services/push-service';
import {
  clearTokens,
  getDeviceId,
  getTokens,
  loadTokens,
  onTokensChange,
  saveTokens,
} from '../services/session-storage';
import { stopSync } from '../services/sync-service';
import { useList } from './use-list';
import { useProfile } from './use-profile';

type SessionStatus = 'loading' | 'signedOut' | 'signedIn';

interface SessionState {
  status: SessionStatus;
  hydrate(): Promise<void>;
  /** Runs login with the device id; an undefined result means the user cancelled. */
  signIn(login: (deviceId: string) => Promise<TokensResponse | undefined>): Promise<boolean>;
  signOut(): Promise<void>;
}

/** Local data belongs to the user: dropped when the session ends, whatever the reason. */
async function wipeLocalData() {
  stopSync();
  useList.getState().clear();
  useProfile.getState().reset();
  await clearActivity();
}

export const useSession = create<SessionState>()((set, get) => ({
  status: 'loading',

  hydrate: async () => {
    const tokens = await loadTokens();
    set({ status: tokens ? 'signedIn' : 'signedOut' });

    // The http client clears the tokens when the refresh token is rejected.
    onTokensChange((next) => {
      if (!next && get().status === 'signedIn') {
        set({ status: 'signedOut' });
        void wipeLocalData();
      }
    });
  },

  signIn: async (login) => {
    const deviceId = await getDeviceId();
    const tokens = await login(deviceId);
    if (!tokens) return false;
    // A previous user may have left data if the app was killed while signing out.
    await clearActivity();
    await saveTokens(tokens);
    set({ status: 'signedIn' });
    return true;
  },

  signOut: async () => {
    const tokens = getTokens();
    // Best effort: the device must stop receiving this user's pushes.
    await disablePush().catch(() => undefined);
    if (tokens) await authApi.logout(tokens.refreshToken).catch(() => undefined);
    set({ status: 'signedOut' });
    await clearTokens();
    await wipeLocalData();
  },
}));
