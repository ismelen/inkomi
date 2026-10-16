import { http } from './http-client';
import { ReleaseResponse } from './types/app';

export const appApi = {
  /** Public: the update check also runs before signing in. */
  latest: () => http.get<ReleaseResponse>('/app/latest', { auth: false }),
};
