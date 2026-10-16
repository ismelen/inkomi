import React from 'react';
import { TextInput, TextInputProps, View } from 'react-native';
import { colors, radii, spacing, typography } from '../../theme';
import { SText } from './s-text';

interface Props extends TextInputProps {
  label?: string;
  hint?: string;
  right?: React.ReactNode;
  left?: React.ReactNode;
}

export function STextInput({ label, hint, right, left, style, ...props }: Props) {
  return (
    <View style={{ gap: spacing.xs }}>
      {!!label && <SText variant="section">{label}</SText>}
      <View
        style={{
          flexDirection: 'row',
          alignItems: 'center',
          gap: spacing.sm,
          borderWidth: 1,
          borderColor: colors.border,
          borderRadius: radii.md,
          paddingHorizontal: spacing.md,
          backgroundColor: colors.background,
        }}
      >
        {left}
        <TextInput
          placeholderTextColor={colors.textSecondary}
          {...props}
          style={[
            typography.body,
            { flex: 1, color: colors.text, paddingVertical: spacing.md - 2 },
            style,
          ]}
        />
        {right}
      </View>
      {!!hint && (
        <SText variant="small" tone="textSecondary">
          {hint}
        </SText>
      )}
    </View>
  );
}
