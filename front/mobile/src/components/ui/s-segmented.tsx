import React from 'react';
import { Pressable, View } from 'react-native';
import { colors, radii, spacing } from '../../theme';
import { SText } from './s-text';

interface Option<T> {
  value: T;
  label: string;
}

interface Props<T> {
  options: Option<T>[];
  value: T;
  onChange: (value: T) => void;
  /** lg: taller, with bigger text, for the main switch of a screen. Default md. */
  size?: 'md' | 'lg';
}

export function SSegmented<T extends string | number | boolean>({
  options,
  value,
  onChange,
  size = 'md',
}: Props<T>) {
  const large = size === 'lg';
  return (
    <View
      style={{
        flexDirection: 'row',
        padding: spacing.xxs,
        borderRadius: radii.md,
        backgroundColor: colors.surface,
        gap: spacing.xxs,
      }}
    >
      {options.map((option) => {
        const selected = option.value === value;
        return (
          <Pressable
            key={String(option.value)}
            onPress={() => onChange(option.value)}
            accessibilityRole="button"
            accessibilityState={{ selected }}
            style={{
              flex: 1,
              paddingVertical: large ? spacing.md : spacing.sm,
              borderRadius: radii.sm,
              alignItems: 'center',
              backgroundColor: selected ? colors.background : 'transparent',
              borderWidth: selected ? 1 : 0,
              borderColor: colors.border,
            }}
          >
            <SText variant={large ? 'title' : 'label'} tone={selected ? 'text' : 'textSecondary'}>
              {option.label}
            </SText>
          </Pressable>
        );
      })}
    </View>
  );
}
