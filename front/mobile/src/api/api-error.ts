export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

/** The request never got a response (offline, timeout, DNS…). */
export class NetworkError extends Error {
  constructor(cause?: unknown) {
    super(cause instanceof Error ? cause.message : 'network error');
    this.name = 'NetworkError';
  }
}

// Exact messages of the backend sentinels → i18n key under `errors.`.
const messageKeys: Record<string, string> = {
  'invalid email': 'invalidEmail',
  'password must have between 8 and 128 characters': 'weakPassword',
  'invalid device id': 'invalidDeviceId',
  'invalid dropbox folder': 'invalidDropboxFolder',
  'email already registered': 'emailTaken',
  'invalid email or password': 'invalidCredentials',
  'invalid or expired refresh token': 'sessionExpired',
  'invalid external token': 'invalidExternalToken',
  'user not found': 'userNotFound',
  'dropbox account not linked': 'dropboxNotLinked',
  'dropbox account linked to another user': 'dropboxAccountTaken',
  'invalid fcm token': 'invalidFcmToken',
  'missing authorization header': 'sessionExpired',
  'invalid authorization format': 'sessionExpired',
  'invalid or expired token': 'sessionExpired',
  'empty query': 'emptyQuery',
  'no libgen mirror available': 'noMirror',
  'source not found': 'sourceNotFound',
  'source not ready': 'sourceNotReady',
  'file not available': 'fileNotAvailable',
  'source not uploaded': 'sourceNotUploaded',
  'e-reader not available': 'ereaderNotAvailable',
  'invalid format': 'invalidRequest',
};

/** i18n key that describes err for the user. */
export function errorKey(err: unknown): string {
  if (err instanceof NetworkError) return 'errors.network';
  if (err instanceof ApiError) {
    const key = messageKeys[err.message];
    if (key) return `errors.${key}`;
  }
  return 'errors.unknown';
}
