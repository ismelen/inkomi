import React, { useRef, useEffect } from 'react';
import { Animated, Pressable } from 'react-native';
import { colors } from '../../theme/colors';

interface Props {
  value: boolean;
  onValueChange: (value: boolean) => void;
}

const WIDTH = 46;
const HEIGHT = 26;
const THUMB_SIZE = 22;
const PADDING = 2;

export default function SSwitch({ value, onValueChange }: Props) {
  const anim = useRef(new Animated.Value(value ? 1 : 0)).current;

  useEffect(() => {
    Animated.timing(anim, {
      toValue: value ? 1 : 0,
      duration: 200,
      useNativeDriver: false,
    }).start();
  }, [value]);

  const translateX = anim.interpolate({
    inputRange: [0, 1],
    outputRange: [PADDING, WIDTH - THUMB_SIZE - PADDING],
  });

  const backgroundColor = anim.interpolate({
    inputRange: [0, 1],
    outputRange: [colors.surface_variant, colors.primary],
  });

  return (
    <Pressable onPress={() => onValueChange(!value)}>
      <Animated.View
        style={{
          width: WIDTH,
          height: HEIGHT,
          borderRadius: HEIGHT / 2,
          backgroundColor,
          justifyContent: 'center',
        }}
      >
        <Animated.View
          style={{
            width: THUMB_SIZE,
            height: THUMB_SIZE,
            borderRadius: THUMB_SIZE / 2,
            backgroundColor: '#fff',
            transform: [{ translateX }],
            shadowColor: '#000',
            shadowOffset: { width: 0, height: 1 },
            shadowOpacity: 0.2,
            shadowRadius: 2,
            elevation: 2,
          }}
        />
      </Animated.View>
    </Pressable>
  );
}
