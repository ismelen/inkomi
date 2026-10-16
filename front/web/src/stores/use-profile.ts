import { create } from 'zustand';
import { usersApi } from '../api/users-api';
import { ProfileResponse } from '../api/types/users';
import { startDropboxAuth } from '../services/dropbox-oauth';

interface ProfileState {
  profile?: ProfileResponse;
  loading: boolean;
  load(): Promise<void>;
  /** Leaves the page for the Dropbox OAuth; the callback page calls completeLink. */
  linkDropbox(): Promise<void>;
  completeLink(refreshToken: string): Promise<void>;
  unlinkDropbox(): Promise<void>;
  setDropboxFolder(folder: string): Promise<void>;
  reset(): void;
}

export const useProfile = create<ProfileState>()((set, get) => ({
  loading: false,

  load: async () => {
    set({ loading: true });
    try {
      set({ profile: await usersApi.me() });
    } finally {
      set({ loading: false });
    }
  },

  linkDropbox: () => startDropboxAuth('link'),

  completeLink: async (refreshToken) => {
    await usersApi.linkDropbox(refreshToken);
    await get().load();
  },

  unlinkDropbox: async () => {
    await usersApi.unlinkDropbox();
    await get().load();
  },

  setDropboxFolder: async (folder) => {
    const res = await usersApi.setDropboxFolder(folder);
    const profile = get().profile;
    if (profile)
      set({ profile: { ...profile, dropbox: { ...profile.dropbox, folder: res.folder } } });
  },

  reset: () => set({ profile: undefined, loading: false }),
}));
