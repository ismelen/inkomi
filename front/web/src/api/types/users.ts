// Mirror of back/auth-service/internal/infra/api/dtos (camelCase).

export type Provider = 'google' | 'dropbox';

export interface DropboxProfileResponse {
  linked: boolean;
  email?: string;
  folder: string;
}

export interface ProfileResponse {
  id: number;
  email?: string;
  emailVerified: boolean;
  hasPassword: boolean;
  providers: Provider[];
  dropbox: DropboxProfileResponse;
}

export interface FCMTokenRequest {
  token: string;
}

export interface DropboxLinkRequest {
  refreshToken: string;
}

export interface DropboxFolderRequest {
  folder: string;
}

export interface DropboxFolderResponse {
  folder: string;
}
