import React, { useEffect, useRef } from 'react';
import { Animated, Pressable } from 'react-native';
import { colors } from '../../theme';

interface Props {
  value: boolean;
  onValueChange: (value: boolean) => void;
  disabled?: boolean;
  accessibilityLabel?: string;
}

const width = 44;
const height = 26;
const thumb = 20;
const padding = 3;

export function SSwitch({ value, onValueChange, disabled, accessibilityLabel }: Props) {
  const anim = useRef(new Animated.Value(value ? 1 : 0)).current;

  useEffect(() => {
    Animated.timing(anim, {
      toValue: value ? 1 : 0,
      duration: 160,
      useNativeDriver: false,
    }).start();
  }, [anim, value]);

  return (
    <Pressable
      onPress={disabled ? undefined : () => onValueChange(!value)}
      accessibilityRole="switch"
      accessibilityState={{ checked: value, disabled }}
      accessibilityLabel={accessibilityLabel}
      hitSlop={8}
      style={{ opacity: disabled ? 0.4 : 1 }}
    >
      <Animated.View
        style={{
          width,
          height,
          borderRadius: height / 2,
          justifyContent: 'center',
          backgroundColor: anim.interpolate({
            inputRange: [0, 1],
            outputRange: [colors.border, colors.primary],
          }),
        }}
      >
        <Animated.View
          style={{
            width: thumb,
            height: thumb,
            borderRadius: thumb / 2,
            backgroundColor: colors.background,
            transform: [
              {
                translateX: anim.interpolate({
                  inputRange: [0, 1],
                  outputRange: [padding, width - thumb - padding],
                }),
              },
            ],
          }}
        />
      </Animated.View>
    </Pressable>
  );
}
