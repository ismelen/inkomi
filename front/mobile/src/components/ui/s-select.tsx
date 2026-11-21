import React, { useState } from 'react';
import { FlatList, Keyboard, Pressable } from 'react-native';
import { colors, radii, spacing } from '../../theme';
import { SIcon } from './s-icon';
import { SSheet } from './s-sheet';
import { SText } from './s-text';

export interface SelectOption<T> {
  value: T;
  label: string;
}

interface Props<T> {
  /** Title of the list and accessibility name; the chip shows the current value. */
  label: string;
  options: SelectOption<T>[];
  value: T;
  onChange: (value: T) => void;
}

/** Compact dropdown: a chip with the current value that opens the options in a sheet. */
export function SSelect<T extends string | number | null>({
  label,
  options,
  value,
  onChange,
}: Props<T>) {
  const [open, setOpen] = useState(false);
  const current = options.find((o) => o.value === value);
  // with the keyboard up the sheet would rise over the screen: close it first
  const openSheet = () => {
    Keyboard.dismiss();
    setOpen(true);
  };

  return (
    <>
      <Pressable
        onPress={openSheet}
        accessibilityRole="button"
        accessibilityLabel={`${label}: ${current?.label ?? ''}`}
        hitSlop={6}
        style={({ pressed }) => ({
          flexDirection: 'row',
          alignItems: 'center',
          gap: spacing.xxs,
          height: 32,
          paddingLeft: spacing.md,
          paddingRight: spacing.sm,
          borderRadius: radii.pill,
          borderWidth: 1,
          borderColor: colors.border,
          backgroundColor: pressed ? colors.surface : colors.background,
        })}
      >
        <SText variant="small" style={{ fontFamily: 'medium' }} numberOfLines={1}>
          {current?.label ?? label}
        </SText>
        <SIcon name="expand_more" size={18} color={colors.textSecondary} />
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
