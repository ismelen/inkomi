import React, { useState } from 'react';
import { StyleSheet, View, ScrollView } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import SText from '../src/components/shared/SText';
import { Stack, usePathname } from 'expo-router';
import { colors } from '../src/theme/colors';
import SourceSelector from '../src/components/senders/source-selector';
import SIcon from '../src/components/icons/SIcon';
import DestinationSelector from '../src/components/senders/destination-selector';
import OptionCardChecker from '../src/components/senders/option-card-checker';
import SButton from '../src/components/shared/SButton';
import SSelect from '../src/components/shared/SSelect';
import { eReaderProfiles } from '../src/constants';
import LoadingScreen from '../src/components/shared/loading-screen';
import { useSender } from '../src/hooks/useSender';
import { useTranslation } from 'react-i18next';
import { useCloud } from '../src/hooks/useCloud';
import SConfirmDialog from '../src/components/shared/SConfirmDialog';
import { useShallow } from 'zustand/react/shallow';

export default function SendBookPage() {
  const insets = useSafeAreaInsets();
  const [showConfirm, setShowConfirm] = useState(false);
  const {
    sending,
    config,
    setConfig,
    send,
    monitoredIdx,
    unmonitor,
    clear: clearSender,
  } = useSender('epub');
  const { t } = useTranslation();
  const { oauth, folder, showAuthConfirm, resolveAuthConfirm } = useCloud(
    useShallow((s) => ({
      oauth: s.oauth,
      folder: s.folder,
      showAuthConfirm: s.showAuthConfirm,
      resolveAuthConfirm: s.resolveAuthConfirm,
    }))
  );

  const isSendDisabled = config.toCloud ? !oauth?.email || !folder : false;

  if (sending)
    return <LoadingScreen title={t('loading.sendingBook')} subtitle={t('loading.optimizing')} />;

  return (
    <>
      <Stack.Screen
        options={{
          headerShown: true,
          title: t('sendBook.title'),
          headerTitleStyle: { fontFamily: 'semibold', fontSize: 20, color: colors.on_background },
          headerTitleAlign: 'center',
          headerStyle: {
            backgroundColor: colors.background,
          },
          headerTintColor: colors.primary,
          headerRight: () => (
            <SButton onPress={clearSender}>
              <SIcon name="delete" size={24} color={colors.primary} type="outlined" />
            </SButton>
          ),
        }}
      />
      <ScrollView style={{ flex: 1, paddingBottom: 24, paddingHorizontal: 24 }}>
        <View style={{ flex: 1, gap: 32, paddingBottom: 24 }}>
          <View style={{ gap: 5 }}>
            <SText style={styles.title}>{t('sendBook.source')}</SText>
            <SourceSelector
              initFolder={config.folder}
              initFiles={config.files}
              onChange={(files, folder) =>
                setConfig((s) => ({ ...s, folder: folder, files: files }))
              }
            />
            {config.folder && (
              <OptionCardChecker
                checked={config.monitoredIdx !== undefined}
                label={t('sendBook.monitorizeFolder')}
                text={t('sendBook.monitorizeText')}
                onChange={(checked) => {
                  if (!checked && monitoredIdx !== undefined) {
                    setShowConfirm(true);
                  } else {
                    setConfig((s) => ({ ...s, monitoredIdx: checked ? 1 : undefined }));
                  }
                }}
              />
            )}
          </View>

          <View>
            <SText style={styles.title}>{t('sendBook.readerModel')}</SText>
            <SSelect
              value={config.model}
              options={eReaderProfiles}
              onOptionChange={(opt) => setConfig((s) => ({ ...s, model: opt.value }))}
            />
          </View>

          <View>
            <SText style={styles.title}>{t('sendBook.destination')}</SText>
            <DestinationSelector
              toCloud={config.toCloud}
              onChange={(toCloud) => setConfig((s) => ({ ...s, toCloud: toCloud }))}
            />
          </View>

          <View style={{ gap: 5 }}>
            <SText style={styles.title}>{t('sendBook.options')}</SText>
            <OptionCardChecker
              checked={config.deleteOriginals ?? false}
              label={t('common.deleteOriginals')}
              text={t('common.deleteOriginalsText')}
              onChange={(checked) => setConfig((s) => ({ ...s, deleteOriginals: checked }))}
            />
          </View>
        </View>
      </ScrollView>

      <SButton
        onPress={() => send()}
        disabled={isSendDisabled}
        style={{
          backgroundColor: isSendDisabled ? colors.surface_variant : colors.primary_container,
          marginTop: 0,
          marginHorizontal: 24,
          marginBottom: 24 + insets.bottom,
          paddingVertical: 12,
          alignItems: 'center',
          justifyContent: 'center',
          borderRadius: 12,
          boxShadow: colors.boxShadow,
        }}
      >
        <SText
          style={{
            fontFamily: 'semibold',
            color: isSendDisabled ? colors.on_surface_variant : colors.on_primary,
          }}
        >
          {t('sendBook.send')}
        </SText>
      </SButton>
      <SConfirmDialog
        visible={showConfirm}
        title={t('common.confirm', 'Confirm')}
        message={t(
          'sendBook.unmonitorConfirm',
          'Are you sure you want to stop monitoring this folder?'
        )}
        confirmText={t('common.ok', 'OK')}
        cancelText={t('common.cancel', 'Cancel')}
        onCancel={() => setShowConfirm(false)}
        onConfirm={() => {
          setShowConfirm(false);
          unmonitor();
        }}
      />
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
    marginBottom: 5,
  },
});
