import React from 'react';
import { Pressable, View } from 'react-native';
import { colors, ColorToken, radii, spacing } from '../../theme';
import { SText } from './s-text';

interface Props {
  label: string;
  selected?: boolean;
  onPress?: () => void;
  /** Colors the border and text, for status chips. */
  tone?: ColorToken;
  filled?: boolean;
  muted?: boolean;
}

export function SChip({ label, selected, onPress, tone, filled, muted }: Props) {
  const color = colors[tone ?? 'text'];
  const background = selected || filled ? color : colors.background;
  const content = (
    <View
      style={{
        paddingHorizontal: spacing.sm + 2,
        paddingVertical: spacing.xxs + 1,
        borderRadius: radii.pill,
        borderWidth: 1,
        borderColor: muted ? colors.surface : selected || filled ? color : colors.border,
        borderStyle: muted ? 'dashed' : 'solid',
        backgroundColor: muted ? colors.surface : background,
      }}
    >
      <SText
        variant="small"
        style={{
          color: muted ? colors.textSecondary : selected || filled ? colors.onPrimary : color,
          fontFamily: 'medium',
        }}
      >
        {label}
      </SText>
    </View>
  );

  if (!onPress) return content;
  return (
    <Pressable
      onPress={onPress}
      accessibilityRole="button"
      accessibilityState={{ selected }}
      hitSlop={4}
    >
      {content}
    </Pressable>
  );
}
