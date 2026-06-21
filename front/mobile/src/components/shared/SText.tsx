import React from 'react';
import { Text, TextProps } from 'react-native';

export default function SText({ children, style, ...props }: TextProps) {
  return (
    <Text
      {...props}
      style={[
        {
          fontFamily: 'regular',
        },
        style,
      ]}
    >
      {children}
    </Text>
  );
}
