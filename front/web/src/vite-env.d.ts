/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_BACKEND_URL?: string;
  readonly VITE_DROPBOX_API_KEY?: string;
  readonly VITE_GOOGLE_WEB_CLIENT_ID?: string;
}

declare const __APP_VERSION__: string;
