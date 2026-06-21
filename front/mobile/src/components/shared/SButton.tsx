import React from 'react';
import { ButtonProps, Pressable, PressableProps, StyleProp, View, ViewStyle } from 'react-native';
import { colors, hexToRgba } from '../../theme/colors';

export default function SButton({ children, style, ...props }: PressableProps) {
  return (
    <Pressable
      {...props}
      android_ripple={{
        color: hexToRgba('#e1e0ff', 0.25),
        borderless: true,
        foreground: true,
      }}
      style={[
        {
          overflow: 'hidden',
        },
        style as StyleProp<ViewStyle>,
      ]}
    >
      {children}
    </Pressable>
  );
}
