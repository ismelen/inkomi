import React, { useState } from 'react';
import { KeyboardAvoidingView, Platform, ScrollView, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { authApi } from '../../src/api/auth-api';
import { errorKey } from '../../src/api/api-error';
import { TokensResponse } from '../../src/api/types/auth';
import { getDropboxRefreshToken } from '../../src/services/dropbox-oauth';
import { getGoogleIdToken } from '../../src/services/google-signin';
import { useSession } from '../../src/stores/use-session';
import { colors, spacing } from '../../src/theme';
import { Screen } from '../../src/components/ui/screen';
import { SButton } from '../../src/components/ui/s-button';
import { SSegmented } from '../../src/components/ui/s-segmented';
import { SText } from '../../src/components/ui/s-text';
import { STextInput } from '../../src/components/ui/s-text-input';

type Mode = 'signIn' | 'register';
type Pending = 'password' | 'google' | 'dropbox' | null;

const minPassword = 8;

export default function LoginScreen() {
  const { t } = useTranslation();
  const signIn = useSession((s) => s.signIn);
  const [mode, setMode] = useState<Mode>('signIn');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [pending, setPending] = useState<Pending>(null);
  const [error, setError] = useState<string | null>(null);

  /** Runs a login; a provider that returns undefined was cancelled by the user. */
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

  const submitPassword = () =>
    run('password', (deviceId) => {
      const body = { email: email.trim(), password, deviceId };
      return mode === 'signIn' ? authApi.login(body) : authApi.register(body);
    });

  const google = () =>
    run('google', async (deviceId) => {
      const idToken = await getGoogleIdToken();
      return idToken ? authApi.google({ idToken, deviceId }) : undefined;
    });

  // Signing in with Dropbox also links it as the upload folder.
  const dropbox = () =>
    run('dropbox', async (deviceId) => {
      const refreshToken = await getDropboxRefreshToken();
      return refreshToken ? authApi.dropbox({ refreshToken, deviceId }) : undefined;
    });

  const canSubmit =
    email.includes('@') &&
    (mode === 'signIn' ? password.length > 0 : password.length >= minPassword);

  return (
    <Screen bottomInset>
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        style={{ flex: 1 }}
      >
        <ScrollView
          keyboardShouldPersistTaps="handled"
          contentContainerStyle={{
            flexGrow: 1,
            justifyContent: 'center',
            padding: spacing.xl,
            gap: spacing.lg,
          }}
        >
          <SText variant="brand" style={{ textAlign: 'center' }}>
            INKOMI
          </SText>
          <SText variant="h1" style={{ textAlign: 'center' }}>
            {t('auth.title')}
          </SText>
          <SSegmented
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
          <STextInput
            label={t('auth.email')}
            placeholder={t('auth.emailPlaceholder')}
            value={email}
            onChangeText={setEmail}
            autoCapitalize="none"
            autoComplete="email"
            keyboardType="email-address"
            textContentType="emailAddress"
          />
          <STextInput
            label={t('auth.password')}
            placeholder="••••••••"
            value={password}
            onChangeText={setPassword}
            secureTextEntry
            autoComplete={mode === 'signIn' ? 'current-password' : 'new-password'}
            hint={mode === 'register' ? t('auth.passwordHint') : undefined}
            onSubmitEditing={canSubmit ? submitPassword : undefined}
          />
          {!!error && (
            <SText tone="danger" style={{ textAlign: 'center' }}>
              {t(error)}
            </SText>
          )}
          <SButton
            label={mode === 'signIn' ? t('auth.signIn') : t('auth.register')}
            onPress={submitPassword}
            disabled={!canSubmit || (pending !== null && pending !== 'password')}
            loading={pending === 'password'}
          />
          <View style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.sm }}>
            <View style={{ flex: 1, height: 1, backgroundColor: colors.border }} />
            <SText variant="small" tone="textSecondary">
              {t('auth.orContinueWith')}
            </SText>
            <View style={{ flex: 1, height: 1, backgroundColor: colors.border }} />
          </View>
          <SButton
            variant="secondary"
            label={t('auth.google')}
            leading={<SText variant="title">G</SText>}
            onPress={google}
            disabled={pending !== null && pending !== 'google'}
            loading={pending === 'google'}
          />
          <SButton
            variant="secondary"
            icon="cloud"
            label={t('auth.dropbox')}
            onPress={dropbox}
            disabled={pending !== null && pending !== 'dropbox'}
            loading={pending === 'dropbox'}
          />
        </ScrollView>
      </KeyboardAvoidingView>
    </Screen>
  );
}
