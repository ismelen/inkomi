import React, { useState } from 'react';
import { KeyboardTypeOptions, StyleProp, View, ViewStyle } from 'react-native';
import { TextInput } from 'react-native-gesture-handler';
import { colors } from '../../theme/colors';
import SText from './SText';

interface Props {
  initial?: string;
  label: string;
  hint: string;
  labelWidth?: number;
  onChange(value: string): void;
  style: StyleProp<ViewStyle>;
  textAlign?: 'center' | 'left' | 'right';
  keyboardType?: KeyboardTypeOptions;
}

export default function STextField({
  initial,
  label,
  hint,
  labelWidth,
  onChange,
  style,
  textAlign,
  keyboardType,
}: Props) {
  const [text, setText] = useState(initial ?? '');

  return (
    <View style={[{ alignItems: 'center', flexDirection: 'row', gap: 14 }, style]}>
      <SText style={{ fontWeight: '500', fontSize: 16, width: labelWidth }}>{label}</SText>
      <TextInput
        style={{ flex: 1, color: colors.onCard, fontSize: 16 }}
        onChangeText={(value) => {
          if (keyboardType === 'number-pad') {
            value = value.replace(/[^0-9]/g, '');
          }
          setText(value);
          onChange(value);
        }}
        textAlign={textAlign}
        keyboardType={keyboardType}
        placeholder={hint}
        placeholderTextColor={colors.gray}
        value={text}
      />
    </View>
  );
}
