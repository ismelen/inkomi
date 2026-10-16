import { create } from 'zustand';
import { authApi } from '../api/auth-api';
import { TokensResponse } from '../api/types/auth';
import { clearActivity } from '../services/activity-db';
import {
  clearTokens,
  getDeviceId,
  getTokens,
  onTokensChange,
  saveTokens,
} from '../services/session-storage';
import { stopSync } from '../services/sync-service';
import { useList } from './use-list';
import { useProfile } from './use-profile';

type SessionStatus = 'signedOut' | 'signedIn';

interface SessionState {
  status: SessionStatus;
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

export const useSession = create<SessionState>()((set) => ({
  status: getTokens() ? 'signedIn' : 'signedOut',

  signIn: async (login) => {
    const tokens = await login(getDeviceId());
    if (!tokens) return false;
    // A previous user may have left data if the tab was closed while signing out.
    await clearActivity();
    saveTokens(tokens);
    set({ status: 'signedIn' });
    return true;
  },

  signOut: async () => {
    const tokens = getTokens();
    if (tokens) await authApi.logout(tokens.refreshToken).catch(() => undefined);
    set({ status: 'signedOut' });
    clearTokens();
    await wipeLocalData();
  },
}));

// The http client clears the tokens when the refresh token is rejected, and other tabs share them.
onTokensChange((next) => {
  const { status } = useSession.getState();
  if (!next && status === 'signedIn') {
    useSession.setState({ status: 'signedOut' });
    void wipeLocalData();
  } else if (next && status === 'signedOut') {
    useSession.setState({ status: 'signedIn' });
  }
});
