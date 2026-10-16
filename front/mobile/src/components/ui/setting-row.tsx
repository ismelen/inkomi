import React from 'react';
import { Pressable, View } from 'react-native';
import { colors, spacing } from '../../theme';
import { SIcon } from './s-icon';
import { SText } from './s-text';

interface Props {
  label: string;
  description?: string;
  /** Right side: a switch, a value… A chevron is shown when there is onPress and no right. */
  right?: React.ReactNode;
  onPress?: () => void;
}

export function SettingRow({ label, description, right, onPress }: Props) {
  return (
    <Pressable
      onPress={onPress}
      disabled={!onPress}
      android_ripple={onPress ? { color: colors.ripple } : undefined}
      style={{
        flexDirection: 'row',
        alignItems: 'center',
        minHeight: 44,
        paddingVertical: spacing.xs,
        gap: spacing.md,
      }}
    >
      <View style={{ flex: 1, gap: spacing.xxs }}>
        <SText variant="body">{label}</SText>
        {!!description && (
          <SText variant="small" tone="textSecondary">
            {description}
          </SText>
        )}
      </View>
      {right ??
        (onPress ? <SIcon name="chevron_right" size={20} color={colors.textSecondary} /> : null)}
    </Pressable>
  );
}

export function SectionTitle({ children }: { children: string }) {
  return (
    <SText variant="section" style={{ marginTop: spacing.lg, marginBottom: spacing.xs }}>
      {children}
    </SText>
  );
}
