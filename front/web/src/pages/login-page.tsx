import { FormEvent, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { authApi } from '../api/auth-api';
import { errorKey } from '../api/api-error';
import { TokensResponse } from '../api/types/auth';
import { Sidebar } from '../components/layout/nav';
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

export function LoginPage({ error: initialError }: { error?: string }) {
  const { t } = useTranslation();
  const signIn = useSession((s) => s.signIn);
  const [mode, setMode] = useState<Mode>('signIn');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
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
    (mode === 'signIn' ? password.length > 0 : password.length >= minPassword);

  return (
    <div className="flex h-dvh">
      <Sidebar disabled />
      <main className="flex flex-1 items-center justify-center overflow-y-auto bg-surface/60 p-4">
        <div className="w-full max-w-[400px] rounded-lg border border-strong bg-bg p-6 md:p-7">
          <div className="mb-5 flex items-center justify-center gap-2.5 md:hidden">
            <span className="size-[26px] rounded-md bg-primary" />
            <span className="text-lg tracking-wide">INKOMI</span>
          </div>
          <h1 className="mb-4 text-center text-2xl">{t('auth.title')}</h1>
          <Segmented
            variant="tabs"
            value={mode}
            onChange={(next) => {
              setMode(next);
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
            <span className="h-px flex-1 bg-surface" />
            {t('auth.orContinueWith')}
            <span className="h-px flex-1 bg-surface" />
          </div>
          <div className="flex flex-col gap-3">
            {googleAvailable() && <GoogleButton onIdToken={google} />}
            <Button
              variant="secondary"
              size="lg"
              loading={pending === 'dropbox'}
              disabled={pending !== null}
              onClick={dropbox}
            >
              <span className="size-3 rotate-45 bg-info" aria-hidden />
              {t('auth.dropbox')}
            </Button>
          </div>
        </div>
      </main>
    </div>
  );
}
