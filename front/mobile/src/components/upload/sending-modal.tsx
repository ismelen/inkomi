import React from 'react';
import { ActivityIndicator, ScrollView } from 'react-native';
import { router } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { colors, spacing } from '../../theme';
import { useUpload } from '../../stores/use-upload';
import { SModal } from '../ui/s-modal';
import { SText } from '../ui/s-text';
import { UploadProgressRow } from './upload-progress-row';

export function SendingModal() {
  const { t } = useTranslation();
  const { visible, phase, files, error, cancel, close } = useUpload();
  const busy = phase === 'requesting' || phase === 'uploading';

  const viewQueue = () => {
    close();
    router.push('/activity');
  };

  return (
    <SModal
      visible={visible}
      persistent={busy}
      title={phase === 'finished' ? t('sending.finished') : t('sending.title')}
      onClose={busy ? cancel : close}
      actions={[
        busy
          ? { label: t('common.cancel'), onPress: cancel, variant: 'secondary' }
          : { label: t('common.close'), onPress: close, variant: 'secondary' },
        { label: t('sending.viewQueue'), onPress: viewQueue, disabled: phase === 'requesting' },
      ]}
    >
      {phase === 'requesting' && (
        <ActivityIndicator color={colors.primary} style={{ marginVertical: spacing.md }} />
      )}
      {!!error && (
        <SText tone="danger" style={{ textAlign: 'center' }}>
          {t(error)}
        </SText>
      )}
      <ScrollView style={{ maxHeight: 320 }} contentContainerStyle={{ gap: spacing.md }}>
        {files.map((file) => (
          <UploadProgressRow key={file.sourceId} file={file} />
        ))}
      </ScrollView>
      {busy && (
        <SText variant="small" tone="textSecondary" style={{ textAlign: 'center' }}>
          {t('sending.keepOpen')}
        </SText>
      )}
    </SModal>
  );
}
