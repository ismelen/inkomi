import Constants from 'expo-constants';

export const backendUrl = (process.env.EXPO_PUBLIC_BACKEND_URL ?? '').replace(/\/+$/, '');
export const wsUrl = backendUrl.replace(/^http/, 'ws') + '/ws';
export const dropboxApiKey = process.env.EXPO_PUBLIC_DROPBOX_API_KEY ?? '';
export const googleWebClientId = process.env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID ?? '';
export const appVersion = Constants.expoConfig?.version ?? '';
