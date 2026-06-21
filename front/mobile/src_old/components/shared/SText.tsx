import React, { ReactNode } from 'react';
import { StyleProp, Text, TextStyle } from 'react-native';
import { colors } from '../../theme/colors';

interface Props {
  children?: ReactNode;
  style?: StyleProp<TextStyle>;
  numberOfLines?: number;
}

export default function SText({ style, children, numberOfLines }: Props) {
  return (
    <Text
      numberOfLines={numberOfLines}
      style={[
        {
          color: colors.white,
        },
        style,
      ]}
    >
      {children}
    </Text>
  );
}
