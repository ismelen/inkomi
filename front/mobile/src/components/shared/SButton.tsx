import React from 'react';
import { ButtonProps, Pressable, PressableProps, StyleProp, View, ViewStyle } from 'react-native';
import { colors, hexToRgba } from '../../theme/colors';

export default function SButton({ children, style, onPress, ...props }: PressableProps) {
  const disabled = props.disabled === true;

  return (
    <Pressable
      {...props}
      android_ripple={{
        color: hexToRgba(colors.primary_fixed, 0.25),
        borderless: !disabled,
        foreground: !disabled,
      }}
      style={[
        {
          overflow: 'hidden',
        },
        style as StyleProp<ViewStyle>,
      ]}
      onPress={(event) => {
        if (disabled) return;
        onPress?.(event);
      }}
    >
      {children}
    </Pressable>
  );
}
