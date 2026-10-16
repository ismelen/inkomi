// Mirror of back/auth-service/internal/infra/api/dtos (camelCase).

export interface LoginRequest {
  email: string;
  password: string;
  deviceId: string;
}

export type RegisterRequest = LoginRequest;

export interface GoogleLoginRequest {
  idToken: string;
  deviceId: string;
}

export interface DropboxLoginRequest {
  refreshToken: string;
  deviceId: string;
}

export interface RefreshTokenRequest {
  refreshToken: string;
}

export interface TokensResponse {
  accessToken: string;
  accessTokenExpiresAt: string;
  refreshToken: string;
  refreshTokenExpiresAt: string;
}
