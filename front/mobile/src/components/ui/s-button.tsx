import React from 'react';
import { ActivityIndicator, Pressable, StyleProp, View, ViewStyle } from 'react-native';
import { colors, radii, spacing } from '../../theme';
import { IconName } from '../icons/icons';
import { SIcon } from './s-icon';
import { SText } from './s-text';

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger';

interface Props {
  label: string;
  onPress?: () => void;
  variant?: Variant;
  icon?: IconName;
  /** Custom leading element (e.g. a brand letter). */
  leading?: React.ReactNode;
  disabled?: boolean;
  loading?: boolean;
  compact?: boolean;
  style?: StyleProp<ViewStyle>;
}

const palette: Record<Variant, { bg: string; fg: string; border: string }> = {
  primary: { bg: colors.primary, fg: colors.onPrimary, border: colors.primary },
  secondary: { bg: colors.background, fg: colors.text, border: colors.border },
  ghost: { bg: 'transparent', fg: colors.text, border: 'transparent' },
  danger: { bg: colors.danger, fg: colors.onPrimary, border: colors.danger },
};

export function SButton({
  label,
  onPress,
  variant = 'primary',
  icon,
  leading,
  disabled,
  loading,
  compact,
  style,
}: Props) {
  const p = palette[variant];
  const inactive = disabled || loading;
  return (
    <Pressable
      onPress={inactive ? undefined : onPress}
      android_ripple={{ color: colors.ripple }}
      accessibilityRole="button"
      accessibilityState={{ disabled: inactive }}
      style={({ pressed }) => [
        {
          minHeight: compact ? 36 : 46,
          paddingHorizontal: compact ? spacing.md : spacing.lg,
          borderRadius: radii.md,
          borderWidth: 1,
          borderColor: p.border,
          backgroundColor: p.bg,
          alignItems: 'center',
          justifyContent: 'center',
          flexDirection: 'row',
          gap: spacing.sm,
          opacity: inactive ? 0.45 : pressed ? 0.85 : 1,
          overflow: 'hidden',
        },
        style,
      ]}
    >
      {loading ? (
        <ActivityIndicator color={p.fg} size="small" />
      ) : (
        <>
          {leading}
          {icon && <SIcon name={icon} size={18} color={p.fg} />}
          <View>
            <SText variant="label" style={{ color: p.fg }}>
              {label}
            </SText>
          </View>
        </>
      )}
    </Pressable>
  );
}
