import React from 'react';
import { Pressable, StyleProp, View, ViewStyle } from 'react-native';
import { colors, radii } from '../../theme';
import { IconName } from '../icons/icons';
import { SIcon } from './s-icon';
import { SText } from './s-text';

interface Props {
  icon: IconName;
  onPress?: () => void;
  label: string;
  color?: string;
  size?: number;
  disabled?: boolean;
  /** Small counter on the top right corner. */
  badge?: number;
  onLongPress?: () => void;
  style?: StyleProp<ViewStyle>;
}

export function SIconButton({
  icon,
  onPress,
  label,
  color = colors.text,
  size = 22,
  disabled,
  badge,
  onLongPress,
  style,
}: Props) {
  return (
    <Pressable
      onPress={disabled ? undefined : onPress}
      onLongPress={onLongPress}
      accessibilityRole="button"
      accessibilityLabel={label}
      hitSlop={6}
      android_ripple={{ color: colors.ripple, borderless: true, radius: size }}
      style={({ pressed }) => [
        {
          width: size + 18,
          height: size + 18,
          borderRadius: radii.pill,
          alignItems: 'center',
          justifyContent: 'center',
          opacity: disabled ? 0.35 : pressed ? 0.6 : 1,
        },
        style,
      ]}
    >
      <SIcon name={icon} size={size} color={color} />
      {!!badge && (
        <View
          style={{
            position: 'absolute',
            top: 2,
            right: 2,
            minWidth: 16,
            height: 16,
            paddingHorizontal: 4,
            borderRadius: radii.pill,
            backgroundColor: colors.warning,
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <SText variant="small" tone="onPrimary" style={{ fontSize: 10, lineHeight: 12 }}>
            {badge > 99 ? '99+' : badge}
          </SText>
        </View>
      )}
    </Pressable>
  );
}
