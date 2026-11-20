import React from 'react';
import { ScrollView } from 'react-native';
import { useTranslation } from 'react-i18next';
import { spacing } from '../../theme';
import { SModal } from '../ui/s-modal';
import { SText } from '../ui/s-text';

interface Props {
  /** Full error of a failed upload; null hides the dialog. */
  error: string | null;
  onClose: () => void;
}

/** The whole error of a failed upload, which the history rows cut to two lines. */
export function ErrorModal({ error, onClose }: Props) {
  const { t } = useTranslation();
  return (
    <SModal
      visible={error !== null}
      title={t('history.errorTitle')}
      onClose={onClose}
      actions={[{ label: t('common.close'), onPress: onClose }]}
    >
      <ScrollView style={{ maxHeight: 360 }} contentContainerStyle={{ paddingBottom: spacing.xs }}>
        <SText variant="small" tone="danger" selectable>
          {error || t('history.errorUnknown')}
        </SText>
      </ScrollView>
    </SModal>
  );
}
