import React, { useState } from 'react';
import { FlatList, Pressable, View } from 'react-native';
import { colors, radii, spacing } from '../../theme';
import { SIcon } from './s-icon';
import { SSheet } from './s-sheet';
import { SText } from './s-text';

export interface SelectOption<T> {
  value: T;
  label: string;
}

interface Props<T> {
  /** Caption above the value, also the title of the list. */
  label: string;
  options: SelectOption<T>[];
  value: T;
  onChange: (value: T) => void;
}

/** Dropdown: a field with the current value that opens the options in a sheet. */
export function SSelect<T extends string | number | null>({
  label,
  options,
  value,
  onChange,
}: Props<T>) {
  const [open, setOpen] = useState(false);
  const current = options.find((o) => o.value === value);

  return (
    <>
      <Pressable
        onPress={() => setOpen(true)}
        accessibilityRole="button"
        accessibilityLabel={`${label}: ${current?.label ?? ''}`}
        android_ripple={{ color: colors.ripple }}
        style={{
          flex: 1,
          flexDirection: 'row',
          alignItems: 'center',
          gap: spacing.xs,
          minHeight: 48,
          paddingHorizontal: spacing.md,
          paddingVertical: spacing.xs,
          borderWidth: 1,
          borderColor: colors.border,
          borderRadius: radii.md,
          overflow: 'hidden',
        }}
      >
        <View style={{ flex: 1 }}>
          <SText variant="small" tone="textSecondary" numberOfLines={1}>
            {label}
          </SText>
          <SText variant="label" numberOfLines={1}>
            {current?.label ?? ''}
          </SText>
        </View>
        <SIcon name="expand_more" size={20} color={colors.textSecondary} />
      </Pressable>
      <SSheet visible={open} title={label} onClose={() => setOpen(false)}>
        <FlatList
          data={options}
          keyExtractor={(o) => String(o.value)}
          renderItem={({ item }) => {
            const selected = item.value === value;
            return (
              <Pressable
                onPress={() => {
                  setOpen(false);
                  if (!selected) onChange(item.value);
                }}
                accessibilityRole="radio"
                accessibilityState={{ selected }}
                android_ripple={{ color: colors.ripple }}
                style={{
                  flexDirection: 'row',
                  alignItems: 'center',
                  gap: spacing.md,
                  minHeight: 52,
                  paddingHorizontal: spacing.sm,
                  borderRadius: radii.md,
                }}
              >
                <SIcon
                  name={selected ? 'radio_button_checked' : 'radio_button_unchecked'}
                  size={22}
                  color={selected ? colors.text : colors.textSecondary}
                />
                <SText variant={selected ? 'title' : 'body'} style={{ flex: 1 }}>
                  {item.label}
                </SText>
              </Pressable>
            );
          }}
        />
      </SSheet>
    </>
  );
}
