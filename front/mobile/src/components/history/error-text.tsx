import React from 'react';
import { Pressable } from 'react-native';
import { useTranslation } from 'react-i18next';
import { SText } from '../ui/s-text';

interface Props {
  /** Error of the upload; empty when the server didn't say why. */
  error: string;
  onPress: (error: string) => void;
}

/** Error line of a history entry, cut to two lines; a tap shows it whole. */
export function ErrorText({ error, onPress }: Props) {
  const { t } = useTranslation();
  return (
    <Pressable
      onPress={() => onPress(error)}
      accessibilityRole="button"
      accessibilityHint={t('history.errorTap')}
      hitSlop={6}
    >
      <SText variant="small" tone="danger" numberOfLines={2}>
        {error ? t('history.error', { error }) : t('history.errorUnknown')}
      </SText>
    </Pressable>
  );
}
