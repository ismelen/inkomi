import React from 'react';
import { ScrollView, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, radii, spacing } from '../../theme';
import { useAppUpdate } from '../../stores/use-app-update';
import { formatBytes } from '../../utils/format';
import { SModal } from '../ui/s-modal';
import { SText } from '../ui/s-text';

export function UpdateModal() {
  const { t } = useTranslation();
  const { release, phase, progress, error, install, dismiss } = useAppUpdate();
  const downloading = phase === 'downloading';

  return (
    <SModal
      visible={!!release && phase !== 'idle'}
      persistent={downloading}
      title={t('update.title', { version: release?.version ?? '' })}
      message={release ? formatBytes(release.size) : undefined}
      onClose={downloading ? () => undefined : dismiss}
      actions={[
        { label: t('update.later'), onPress: dismiss, variant: 'secondary', disabled: downloading },
        {
          label: phase === 'error' ? t('common.retry') : t('update.install'),
          onPress: () => void install(),
          loading: downloading,
        },
      ]}
    >
      {!!release?.notes && (
        <ScrollView style={{ maxHeight: 200 }}>
          <SText tone="textSecondary">{release.notes}</SText>
        </ScrollView>
      )}
      {downloading && (
        <View style={{ gap: spacing.xs }}>
          <View style={{ height: 4, borderRadius: radii.pill, backgroundColor: colors.surface }}>
            <View
              style={{
                width: `${Math.round(progress * 100)}%`,
                height: 4,
                borderRadius: radii.pill,
                backgroundColor: colors.primary,
              }}
            />
          </View>
          <SText variant="small" tone="textSecondary" style={{ textAlign: 'center' }}>
            {t('update.downloading', { percent: Math.round(progress * 100) })}
          </SText>
        </View>
      )}
      {phase === 'error' && !!error && (
        <SText tone="danger" style={{ textAlign: 'center' }}>
          {t(error)}
        </SText>
      )}
    </SModal>
  );
}
