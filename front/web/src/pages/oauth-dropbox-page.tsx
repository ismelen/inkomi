import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router';
import { useTranslation } from 'react-i18next';
import { authApi } from '../api/auth-api';
import { errorKey } from '../api/api-error';
import { Spinner } from '../components/ui/spinner';
import { completeDropboxAuth } from '../services/dropbox-oauth';
import { useProfile } from '../stores/use-profile';
import { useSession } from '../stores/use-session';
import { showError } from '../stores/use-toast';
import { LoginPage } from './login-page';

/** Redirect target of the Dropbox OAuth: signs in or links the account, then goes back. */
export function OAuthDropboxPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [loginError, setLoginError] = useState<string | null>(null);
  const started = useRef(false);

  useEffect(() => {
    // Strict mode runs effects twice; the code can only be exchanged once.
    if (started.current) return;
    started.current = true;

    const finish = async () => {
      const result = await completeDropboxAuth(window.location.search).catch((err) => {
        setLoginError(errorKey(err));
        return undefined;
      });
      if (result === undefined) return;
      if (!result) return navigate('/', { replace: true });

      const { refreshToken, intent, returnTo } = result;
      if (intent === 'login') {
        if (!refreshToken) return navigate('/', { replace: true });
        try {
          await useSession
            .getState()
            .signIn((deviceId) => authApi.dropbox({ refreshToken, deviceId }));
          navigate('/', { replace: true });
        } catch (err) {
          setLoginError(errorKey(err));
        }
        return;
      }

      if (refreshToken) {
        try {
          await useProfile.getState().completeLink(refreshToken);
        } catch (err) {
          showError(err);
        }
      }
      navigate(returnTo || '/', { replace: true });
    };
    void finish();
  }, [navigate]);

  if (loginError) return <LoginPage error={loginError} />;

  return (
    <div className="flex h-dvh items-center justify-center gap-3 text-base text-muted">
      <Spinner size={20} /> {t('oauth.connecting')}
    </div>
  );
}
