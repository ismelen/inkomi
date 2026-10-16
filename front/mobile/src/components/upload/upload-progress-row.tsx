import React from 'react';
import { View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, radii, spacing } from '../../theme';
import { UploadFile } from '../../stores/use-upload';
import { SText } from '../ui/s-text';

export function UploadProgressRow({ file }: { file: UploadFile }) {
  const { t } = useTranslation();
  const status = {
    waiting: { text: t('sending.waiting'), tone: 'textSecondary' as const },
    uploading: { text: `${Math.round(file.progress * 100)}%`, tone: 'text' as const },
    uploaded: { text: t('sending.uploaded'), tone: 'success' as const },
    failed: { text: t('sending.failed'), tone: 'danger' as const },
  }[file.state];

  return (
    <View style={{ gap: spacing.xs }}>
      <View style={{ flexDirection: 'row', gap: spacing.sm }}>
        <SText variant="body" numberOfLines={1} style={{ flex: 1 }}>
          {file.title}
        </SText>
        <SText variant="small" tone={status.tone}>
          {status.text}
        </SText>
      </View>
      {file.state === 'uploading' && (
        <View style={{ height: 3, borderRadius: radii.pill, backgroundColor: colors.surface }}>
          <View
            style={{
              width: `${Math.round(file.progress * 100)}%`,
              height: 3,
              borderRadius: radii.pill,
              backgroundColor: colors.primary,
            }}
          />
        </View>
      )}
      {file.state === 'failed' && !!file.error && (
        <SText variant="small" tone="danger">
          {t(file.error)}
        </SText>
      )}
    </View>
  );
}
