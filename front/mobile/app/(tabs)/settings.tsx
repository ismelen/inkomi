import React, { useState } from 'react';
import { StyleSheet, View, ActivityIndicator } from 'react-native';
import { ScrollView } from 'react-native-gesture-handler';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import SText from '../../src/components/shared/SText';
import { colors } from '../../src/theme/colors';
import SSelect from '../../src/components/shared/SSelect';
import { eReaderProfiles } from '../../src/constants';
import { useSettings } from '../../src/hooks/useSettings';
import { useShallow } from 'zustand/react/shallow';
import SIcon from '../../src/components/icons/SIcon';
import { useCloud } from '../../src/hooks/useCloud';
import SButton from '../../src/components/shared/SButton';
import { usePathname } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { SupportedLanguage } from '../../src/i18n/i18n';
import SConfirmDialog from '../../src/components/shared/SConfirmDialog';
import CloudConfig from '../../src/components/shared/cloud-config';

const languageOptions: { value: SupportedLanguage; label: string }[] = [
  { value: 'en', label: 'English' },
  { value: 'es', label: 'Español' },
];

export default function SettingsPage() {
  const insets = useSafeAreaInsets();
  const pathname = usePathname();
  const { t, i18n } = useTranslation();

  const { settings, setModel, setLanguage } = useSettings(
    useShallow((s) => ({ settings: s.settings, setModel: s.setModel, setLanguage: s.setLanguage }))
  );

  const { oauth, folder, getFolder, getToken, logout, showAuthConfirm, resolveAuthConfirm } =
    useCloud(
      useShallow((s) => ({
        oauth: s.oauth,
        folder: s.folder,
        getToken: s.getToken,
        getFolder: s.getFolder,
        logout: s.logout,
        showAuthConfirm: s.showAuthConfirm,
        resolveAuthConfirm: s.resolveAuthConfirm,
      }))
    );

  const [loading, setLoading] = useState(false);

  return (
    <>
      <ScrollView
        style={{ flex: 1, paddingHorizontal: 24 }}
        contentContainerStyle={{ paddingBottom: 24 + insets.bottom }}
      >
        <SText style={{ fontFamily: 'bold', fontSize: 28 }}>{t('settings.title')}</SText>
        <View style={{ marginTop: 14, gap: 4 }}>
          <SText style={styles.title}>{t('settings.readerModel')}</SText>
          <SSelect
            value={settings.model}
            options={eReaderProfiles}
            onOptionChange={(opt) => setModel(opt.value)}
          />
        </View>

        <View style={{ marginTop: 32, gap: 4 }}>
          <SText style={styles.title}>{t('settings.language')}</SText>
          <SSelect
            value={settings.language ?? (i18n.resolvedLanguage || i18n.language || 'en')}
            options={languageOptions}
            onOptionChange={(opt) => setLanguage(opt.value as SupportedLanguage)}
          />
        </View>

        <View style={{ marginTop: 32, gap: 4 }}>
          <SText style={styles.title}>{t('settings.cloudSync')}</SText>
          <CloudConfig />
        </View>
      </ScrollView>
      <SConfirmDialog
        visible={showAuthConfirm}
        title={t('common.authConfirmTitle')}
        message={t('common.authConfirmMessage')}
        confirmText={t('common.authConfirmOk')}
        cancelText={t('common.cancel')}
        onCancel={() => resolveAuthConfirm(false)}
        onConfirm={() => resolveAuthConfirm(true)}
      />
    </>
  );
}

const styles = StyleSheet.create({
  title: {
    fontFamily: 'semibold',
    fontSize: 14,
    color: colors.on_surface_variant,
  },
});
