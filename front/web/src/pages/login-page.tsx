import { FormEvent, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { authApi } from '../api/auth-api';
import { errorKey } from '../api/api-error';
import { TokensResponse } from '../api/types/auth';
import { Brand, Sidebar } from '../components/layout/nav';
import { Button } from '../components/ui/button';
import { TextInput } from '../components/ui/field';
import { Segmented } from '../components/ui/segmented';
import { startDropboxAuth } from '../services/dropbox-oauth';
import { googleAvailable, renderGoogleButton } from '../services/google-signin';
import { useSession } from '../stores/use-session';

type Mode = 'signIn' | 'register';
type Pending = 'password' | 'google' | 'dropbox' | null;

const minPassword = 8;

function GoogleButton({ onIdToken }: { onIdToken: (idToken: string) => void }) {
  const { i18n } = useTranslation();
  const host = useRef<HTMLDivElement>(null);
  const callback = useRef(onIdToken);
  callback.current = onIdToken;

  useEffect(() => {
    const el = host.current;
    if (!el) return;
    const dark =
      document.documentElement.dataset.theme === 'dark' ||
      (!document.documentElement.dataset.theme &&
        window.matchMedia('(prefers-color-scheme: dark)').matches);
    renderGoogleButton(
      el,
      { width: Math.min(400, el.clientWidth), locale: i18n.language, dark },
      (token) => callback.current(token)
    ).catch((err) => console.warn('google', err));
  }, [i18n.language]);

  return <div ref={host} className="flex h-10 justify-center overflow-hidden" />;
}

/** Google's "G" (its brand colors), for the placeholder when Google isn't configured. */
function GoogleLogo() {
  return (
    <svg viewBox="0 0 48 48" width={20} height={20} aria-hidden className="shrink-0">
      <path
        fill="#EA4335"
        d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"
      />
      <path
        fill="#4285F4"
        d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"
      />
      <path
        fill="#FBBC05"
        d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z"
      />
      <path
        fill="#34A853"
        d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"
      />
    </svg>
  );
}

/** Dropbox's mark (simple-icons), in its brand blue. */
function DropboxLogo() {
  return (
    <svg viewBox="0 0 24 24" width={20} height={20} aria-hidden className="shrink-0 fill-dropbox">
      <path d="M6 1.807L0 5.629l6 3.822 6.001-3.822L6 1.807zM18 1.807l-6 3.822 6 3.822 6-3.822-6-3.822zM0 13.274l6 3.822 6.001-3.822L6 9.452l-6 3.822zM18 9.452l-6 3.822 6 3.822 6-3.822-6-3.822zM6 18.371l6.001 3.822 6-3.822-6-3.822L6 18.371z" />
    </svg>
  );
}

export function LoginPage({ error: initialError }: { error?: string }) {
  const { t } = useTranslation();
  const signIn = useSession((s) => s.signIn);
  const [mode, setMode] = useState<Mode>('signIn');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [repeat, setRepeat] = useState('');
  const [pending, setPending] = useState<Pending>(null);
  const [error, setError] = useState<string | null>(initialError ?? null);

  const run = async (
    kind: Exclude<Pending, null>,
    login: (deviceId: string) => Promise<TokensResponse | undefined>
  ) => {
    setPending(kind);
    setError(null);
    try {
      await signIn(login);
    } catch (err) {
      setError(errorKey(err));
    } finally {
      setPending(null);
    }
  };

  const submit = (e: FormEvent) => {
    e.preventDefault();
    void run('password', (deviceId) => {
      const body = { email: email.trim(), password, deviceId };
      return mode === 'signIn' ? authApi.login(body) : authApi.register(body);
    });
  };

  const google = (idToken: string) =>
    void run('google', (deviceId) => authApi.google({ idToken, deviceId }));

  // Signing in with Dropbox also links it as the upload folder. The page leaves for Dropbox.
  const dropbox = () => {
    setPending('dropbox');
    startDropboxAuth('login').catch((err) => {
      setError(errorKey(err));
      setPending(null);
    });
  };

  const canSubmit =
    email.includes('@') &&
    (mode === 'signIn'
      ? password.length > 0
      : password.length >= minPassword && repeat === password);
  const mismatch = mode === 'register' && repeat.length > 0 && repeat !== password;

  return (
    <div className="flex h-dvh">
      <Sidebar disabled />
      <main className="flex flex-1 items-center justify-center overflow-y-auto p-4">
        <div className="w-full max-w-[420px] rounded-xl bg-card p-6 shadow-modal md:p-8">
          <Brand className="mb-6 justify-center md:hidden" nameClassName="text-2xl" />
          <h1 className="mb-5 text-center text-2xl font-bold">{t('auth.title')}</h1>
          <Segmented
            variant="tabs"
            value={mode}
            onChange={(next) => {
              setMode(next);
              setRepeat('');
              setError(null);
            }}
            options={[
              { value: 'signIn', label: t('auth.signInTab') },
              { value: 'register', label: t('auth.registerTab') },
            ]}
          />
          <form onSubmit={submit} className="mt-6 flex flex-col gap-4">
            <TextInput
              label={t('auth.email')}
              type="email"
              autoComplete="email"
              placeholder={t('auth.emailPlaceholder')}
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
            <TextInput
              label={t('auth.password')}
              type="password"
              autoComplete={mode === 'signIn' ? 'current-password' : 'new-password'}
              placeholder="••••••••"
              value={password}
              hint={mode === 'register' ? t('auth.passwordHint') : undefined}
              onChange={(e) => setPassword(e.target.value)}
            />
            {mode === 'register' && (
              <TextInput
                label={t('auth.repeatPassword')}
                type="password"
                autoComplete="new-password"
                placeholder="••••••••"
                value={repeat}
                error={mismatch ? t('auth.passwordMismatch') : undefined}
                onChange={(e) => setRepeat(e.target.value)}
              />
            )}
            {error && (
              <p role="alert" className="text-sm text-danger">
                {t(error)}
              </p>
            )}
            <Button
              type="submit"
              size="lg"
              className="mt-1"
              disabled={!canSubmit || pending !== null}
              loading={pending === 'password'}
            >
              {mode === 'signIn' ? t('auth.signIn') : t('auth.register')}
            </Button>
          </form>
          <div className="my-5 flex items-center gap-3 text-sm text-muted">
            <span className="h-px flex-1 bg-variant" />
            {t('auth.orContinueWith')}
            <span className="h-px flex-1 bg-variant" />
          </div>
          <div className="flex flex-col gap-3">
            {googleAvailable() ? (
              <GoogleButton onIdToken={google} />
            ) : (
              // Without VITE_GOOGLE_WEB_CLIENT_ID (local, mock) there's no Google button to render.
              <Button variant="secondary" size="lg" disabled title={t('auth.googleUnavailable')}>
                <GoogleLogo />
                {t('auth.google')}
              </Button>
            )}
            <Button
              variant="secondary"
              size="lg"
              loading={pending === 'dropbox'}
              disabled={pending !== null}
              onClick={dropbox}
            >
              <DropboxLogo />
              {t('auth.dropbox')}
            </Button>
          </div>
        </div>
      </main>
    </div>
  );
}
