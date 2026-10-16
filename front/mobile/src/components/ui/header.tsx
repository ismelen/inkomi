import React from 'react';
import { View } from 'react-native';
import { router } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { spacing } from '../../theme';
import { SIconButton } from './s-icon-button';
import { SText } from './s-text';

interface Props {
  title: string;
  /** Shows the back chevron. Default true. */
  back?: boolean;
  right?: React.ReactNode;
}

export function Header({ title, back = true, right }: Props) {
  const { t } = useTranslation();
  return (
    <View
      style={{
        height: 52,
        flexDirection: 'row',
        alignItems: 'center',
        paddingHorizontal: spacing.sm,
        gap: spacing.xs,
      }}
    >
      {back ? (
        <SIconButton icon="chevron_left" label={t('common.back')} onPress={() => router.back()} />
      ) : (
        <View style={{ width: spacing.sm }} />
      )}
      <SText variant="title" numberOfLines={1} style={{ flex: 1 }}>
        {title}
      </SText>
      {right}
    </View>
  );
}
