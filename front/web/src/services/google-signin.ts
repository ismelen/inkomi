import { googleWebClientId } from '../config';

interface CredentialResponse {
  credential?: string;
}

interface GoogleAccountsId {
  initialize(config: {
    client_id: string;
    callback: (response: CredentialResponse) => void;
    ux_mode?: 'popup' | 'redirect';
  }): void;
  renderButton(
    parent: HTMLElement,
    options: {
      type?: 'standard' | 'icon';
      theme?: 'outline' | 'filled_black';
      size?: 'large' | 'medium';
      text?: 'continue_with' | 'signin_with';
      shape?: 'rectangular' | 'pill';
      width?: number;
      locale?: string;
    }
  ): void;
}

declare global {
  interface Window {
    google?: { accounts: { id: GoogleAccountsId } };
  }
}

let scriptPromise: Promise<GoogleAccountsId> | null = null;

function loadScript(): Promise<GoogleAccountsId> {
  scriptPromise ??= new Promise((resolve, reject) => {
    const script = document.createElement('script');
    script.src = 'https://accounts.google.com/gsi/client';
    script.async = true;
    script.onload = () =>
      window.google ? resolve(window.google.accounts.id) : reject(new Error('gsi unavailable'));
    script.onerror = () => {
      scriptPromise = null;
      reject(new Error('gsi unavailable'));
    };
    document.head.appendChild(script);
  });
  return scriptPromise;
}

export function googleAvailable(): boolean {
  return !!googleWebClientId;
}

/**
 * Renders the Google Identity Services button into parent. Its popup returns an id token for
 * POST /auth/google (issued to the web client id, accepted by the auth-service).
 */
export async function renderGoogleButton(
  parent: HTMLElement,
  opts: { width: number; locale: string; dark: boolean },
  onIdToken: (idToken: string) => void
): Promise<void> {
  const gsi = await loadScript();
  gsi.initialize({
    client_id: googleWebClientId,
    ux_mode: 'popup',
    callback: (response) => {
      if (response.credential) onIdToken(response.credential);
    },
  });
  parent.replaceChildren();
  gsi.renderButton(parent, {
    type: 'standard',
    theme: opts.dark ? 'filled_black' : 'outline',
    size: 'large',
    text: 'continue_with',
    shape: 'rectangular',
    width: opts.width,
    locale: opts.locale,
  });
}
