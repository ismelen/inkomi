import React from 'react';
import { View } from 'react-native';
import Slider from '@react-native-community/slider';
import { colors, spacing } from '../../theme';
import { Range } from '../../models/manga-config';
import { formatNumber } from '../../utils/format';
import { SText } from './s-text';

interface Props {
  label: string;
  value: number;
  range: Range;
  onChange: (value: number) => void;
  format?: (value: number) => string;
  disabled?: boolean;
  /** Element at the right of the label, e.g. a switch that enables the value. */
  accessory?: React.ReactNode;
}

export function SSlider({ label, value, range, onChange, format, disabled, accessory }: Props) {
  const text = format ? format(value) : formatNumber(value, range.step);
  return (
    <View style={{ paddingVertical: spacing.xs, opacity: disabled ? 0.45 : 1 }}>
      <View style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.sm }}>
        <SText variant="body" style={{ flex: 1 }}>
          {label}
        </SText>
        {accessory}
        <SText variant="label" style={{ minWidth: 44, textAlign: 'right' }}>
          {disabled ? '—' : text}
        </SText>
      </View>
      <Slider
        value={value}
        minimumValue={range.min}
        maximumValue={range.max}
        step={range.step}
        disabled={disabled}
        onValueChange={(v) => onChange(Number(formatNumber(v, range.step)))}
        minimumTrackTintColor={colors.primary}
        maximumTrackTintColor={colors.border}
        thumbTintColor={colors.primary}
        style={{ marginHorizontal: -spacing.sm, height: 32 }}
        accessibilityLabel={label}
      />
    </View>
  );
}
