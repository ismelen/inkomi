import { http } from './http-client';
import {
  DropboxFolderEntry,
  DropboxFolderResponse,
  DropboxFoldersResponse,
  ProfileResponse,
} from './types/users';

export const usersApi = {
  me: () => http.get<ProfileResponse>('/users/me'),
  /** An empty token stops the push notifications of the user. */
  setFcmToken: (token: string) => http.put<void>('/users/me/fcm-token', { token }),
  linkDropbox: (refreshToken: string) => http.put<void>('/users/me/dropbox', { refreshToken }),
  unlinkDropbox: () => http.delete<void>('/users/me/dropbox'),
  setDropboxFolder: (folder: string) =>
    http.put<DropboxFolderResponse>('/users/me/dropbox/folder', { folder }),
  /** Subfolders of path in the user's Dropbox, for the folder picker. */
  listDropboxFolders: (path: string) =>
    http.get<DropboxFoldersResponse>('/users/me/dropbox/folders', { query: { path } }),
  createDropboxFolder: (path: string) =>
    http.post<DropboxFolderEntry>('/users/me/dropbox/folders', { path }),
};
