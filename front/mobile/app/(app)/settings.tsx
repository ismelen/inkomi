import React, { useEffect, useState } from 'react';
import { ActivityIndicator, Alert, ScrollView, View } from 'react-native';
import { router } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { errorKey } from '../../src/api/api-error';
import { EReaderPicker } from '../../src/components/settings/ereader-picker';
import { Header } from '../../src/components/ui/header';
import { Screen } from '../../src/components/ui/screen';
import { SButton } from '../../src/components/ui/s-button';
import { SSegmented } from '../../src/components/ui/s-segmented';
import { SSwitch } from '../../src/components/ui/s-switch';
import { SText } from '../../src/components/ui/s-text';
import { STextInput } from '../../src/components/ui/s-text-input';
import { SectionTitle, SettingRow } from '../../src/components/ui/setting-row';
import { appVersion } from '../../src/config';
import { supportedLanguages } from '../../src/i18n/i18n';
import { findEReader } from '../../src/models/ereaders';
import { disablePush, enablePush } from '../../src/services/push-service';
import { useAppUpdate } from '../../src/stores/use-app-update';
import { useProfile } from '../../src/stores/use-profile';
import { useSession } from '../../src/stores/use-session';
import { useSettings } from '../../src/stores/use-settings';
import { colors, spacing } from '../../src/theme';

export default function SettingsScreen() {
  const { t, i18n } = useTranslation();
  const signOut = useSession((s) => s.signOut);
  const { profile, loading, load, linkDropbox, unlinkDropbox, setDropboxFolder } = useProfile();
  const settings = useSettings();
  const ereader = findEReader(settings.ereaderKey);
  const [pickerOpen, setPickerOpen] = useState(false);
  const [folder, setFolder] = useState(profile?.dropbox.folder ?? '');
  const [busy, setBusy] = useState<string | null>(null);

  useEffect(() => {
    setFolder(profile?.dropbox.folder ?? '');
  }, [profile?.dropbox.folder]);

  const attempt = async (key: string, action: () => Promise<unknown>) => {
    setBusy(key);
    try {
      await action();
    } catch (err) {
      Alert.alert(t('errors.title'), t(errorKey(err)));
    } finally {
      setBusy(null);
    }
  };

  const saveFolder = () => {
    const next = folder.trim();
    if (!profile || next === profile.dropbox.folder) return;
    void attempt('folder', async () => {
      try {
        await setDropboxFolder(next);
      } catch (err) {
        setFolder(profile.dropbox.folder);
        throw err;
      }
    });
  };

  const toggleNotifications = (value: boolean) =>
    attempt('notifications', async () => {
      if (!value) {
        await disablePush();
        settings.setNotifications(false);
        return;
      }
      if (await enablePush()) settings.setNotifications(true);
      else Alert.alert(t('settings.notifications'), t('settings.notificationsDenied'));
    });

  const dropbox = profile?.dropbox;

  return (
    <Screen bottomInset>
      <Header title={t('settings.title')} />
      <ScrollView
        contentContainerStyle={{ paddingHorizontal: spacing.lg, paddingBottom: spacing.xxl }}
      >
        <SectionTitle>{t('settings.account')}</SectionTitle>
        <SettingRow
          label={profile?.email ?? (loading ? '…' : t('settings.noEmail'))}
          right={
            <SButton
              variant="secondary"
              compact
              icon="logout"
              label={t('settings.signOut')}
              loading={busy === 'signOut'}
              onPress={() => void attempt('signOut', signOut)}
            />
          }
        />

        <SectionTitle>{t('settings.reader')}</SectionTitle>
        <SettingRow
          label={ereader.name}
          description={t(ereader.kepub ? 'settings.readerKepub' : 'settings.readerSize', {
            width: ereader.width,
            height: ereader.height,
          })}
          onPress={() => setPickerOpen(true)}
        />

        <SectionTitle>{t('settings.dropbox')}</SectionTitle>
        {!profile ? (
          loading ? (
            <ActivityIndicator color={colors.primary} />
          ) : (
            <SButton
              variant="secondary"
              compact
              label={t('common.retry')}
              onPress={() => void attempt('profile', load)}
            />
          )
        ) : dropbox?.linked ? (
          <>
            <SettingRow
              label={t('settings.connected')}
              description={dropbox.email}
              right={
                <SButton
                  variant="secondary"
                  compact
                  label={t('settings.disconnect')}
                  loading={busy === 'unlink'}
                  onPress={() => void attempt('unlink', unlinkDropbox)}
                />
              }
            />
            <View style={{ marginVertical: spacing.sm }}>
              <STextInput
                label={t('settings.folder')}
                value={folder}
                onChangeText={setFolder}
                onBlur={saveFolder}
                onSubmitEditing={saveFolder}
                autoCapitalize="none"
                autoCorrect={false}
                right={busy === 'folder' ? <ActivityIndicator color={colors.primary} /> : undefined}
              />
            </View>
            <SettingRow
              label={t('settings.toCloudDefault')}
              right={
                <SSwitch
                  value={settings.toCloudDefault}
                  onValueChange={settings.setToCloudDefault}
                />
              }
            />
          </>
        ) : (
          <SettingRow
            label={t('settings.disconnected')}
            right={
              <SButton
                compact
                icon="cloud"
                label={t('settings.connect')}
                loading={busy === 'link'}
                onPress={() => void attempt('link', linkDropbox)}
              />
            }
          />
        )}

        <SectionTitle>{t('settings.notifications')}</SectionTitle>
        <SettingRow
          label={t('settings.notifications')}
          right={
            <SSwitch
              value={settings.notifications}
              disabled={busy === 'notifications'}
              onValueChange={(v) => void toggleNotifications(v)}
            />
          }
        />

        <SectionTitle>{t('settings.language')}</SectionTitle>
        <SSegmented
          value={(settings.language ?? i18n.language) as (typeof supportedLanguages)[number]}
          onChange={(lang) => void settings.setLanguage(lang)}
          options={supportedLanguages.map((lang) => ({
            value: lang,
            label: t(`languages.${lang}`),
          }))}
        />

        <SectionTitle>{t('settings.defaultManga')}</SectionTitle>
        <SettingRow
          label={t('settings.defaultManga')}
          onPress={() => router.push({ pathname: '/options/[id]', params: { id: 'default' } })}
        />

        <SectionTitle>{t('update.section')}</SectionTitle>
        <SettingRow
          label={t('update.check')}
          description={t('settings.version', { version: appVersion })}
          right={busy === 'update' ? <ActivityIndicator color={colors.primary} /> : undefined}
          onPress={() =>
            void attempt('update', async () => {
              if (!(await useAppUpdate.getState().check(true))) {
                Alert.alert(t('update.section'), t('update.upToDate'));
              }
            })
          }
        />
      </ScrollView>
      <EReaderPicker
        visible={pickerOpen}
        title={t('settings.reader')}
        value={settings.ereaderKey}
        onChange={settings.setEReader}
        onClose={() => setPickerOpen(false)}
      />
    </Screen>
  );
}
