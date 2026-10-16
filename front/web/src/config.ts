export const backendUrl = (import.meta.env.VITE_BACKEND_URL ?? '').replace(/\/+$/, '');
export const wsUrl = (backendUrl || window.location.origin).replace(/^http/, 'ws') + '/ws';
export const dropboxApiKey = import.meta.env.VITE_DROPBOX_API_KEY ?? '';
export const googleWebClientId = import.meta.env.VITE_GOOGLE_WEB_CLIENT_ID ?? '';
export const appVersion = __APP_VERSION__;
