import React, { useState } from 'react';
import { View } from 'react-native';
import { Switch } from 'react-native-switch';
import { colors } from '../../theme/colors';
import SText from '../shared/SText';

interface Props {
  initial: boolean;
  onChange(value: boolean): void;
  label: string;
}

export default function ConfigToggleField({ initial, onChange, label }: Props) {
  const [checked, setChecked] = useState(initial);

  return (
    <View
      style={{
        paddingHorizontal: 16,
        paddingVertical: 10,
        flexDirection: 'row',
        alignItems: 'center',
      }}
    >
      <SText style={{ flex: 1, fontWeight: '500', fontSize: 16 }}>{label}</SText>
      <Switch
        value={checked}
        onValueChange={(value) => {
          setChecked(value);
          onChange(value);
        }}
        renderActiveText={false}
        renderInActiveText={false}
        backgroundInactive={colors.gray}
        backgroundActive={colors.primary}
        circleBorderActiveColor={colors.white}
        circleBorderInactiveColor={colors.white}
        switchWidthMultiplier={1.5}
        barHeight={27}
        circleSize={25}
      />
    </View>
  );
}
