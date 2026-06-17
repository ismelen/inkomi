import React, { ReactNode } from 'react';
import { Pressable, StyleProp, StyleSheet, ViewStyle } from 'react-native';
import { colors } from '../../theme/colors';

interface Props {
  style?: StyleProp<ViewStyle>;
  children: ReactNode;
  onPress?(): void;
}

export default function SButton({ children, style, onPress }: Props) {
  return (
    <Pressable onPress={onPress} style={[styles.button, style]}>
      {children}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  button: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: colors.primary,
    borderRadius: 14,
    paddingHorizontal: 16,
    paddingVertical: 14,
  },
});
