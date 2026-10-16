import { http } from './http-client';
import {
  DropboxLoginRequest,
  GoogleLoginRequest,
  LoginRequest,
  RegisterRequest,
  TokensResponse,
} from './types/auth';

const noAuth = { auth: false };

export const authApi = {
  register: (body: RegisterRequest) => http.post<TokensResponse>('/auth/register', body, noAuth),
  login: (body: LoginRequest) => http.post<TokensResponse>('/auth/login', body, noAuth),
  google: (body: GoogleLoginRequest) => http.post<TokensResponse>('/auth/google', body, noAuth),
  dropbox: (body: DropboxLoginRequest) => http.post<TokensResponse>('/auth/dropbox', body, noAuth),
  logout: (refreshToken: string) => http.post<void>('/auth/logout', { refreshToken }, noAuth),
};
