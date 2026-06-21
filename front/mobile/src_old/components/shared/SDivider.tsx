import React from 'react';
import { View } from 'react-native';
import { colors } from '../../theme/colors';

export default function SDivider() {
  return (
    <View
      style={{
        backgroundColor: colors.gray,
        height: 0.4,
      }}
    />
  );
}
