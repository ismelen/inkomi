import React from 'react';
import { Text, TextProps } from 'react-native';
import { colors, ColorToken, typography, TypographyVariant } from '../../theme';

interface Props extends TextProps {
  variant?: TypographyVariant;
  tone?: ColorToken;
}

export function SText({ variant = 'body', tone, style, children, ...props }: Props) {
  const section = variant === 'section';
  return (
    <Text
      {...props}
      style={[
        typography[variant],
        { color: colors[tone ?? (section ? 'textSecondary' : 'text')] },
        section && { textTransform: 'uppercase' },
        style,
      ]}
    >
      {children}
    </Text>
  );
}
