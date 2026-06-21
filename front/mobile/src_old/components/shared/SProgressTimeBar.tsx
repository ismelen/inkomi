import React, { useEffect, useRef } from 'react';
import { Animated, StyleProp, View, ViewStyle } from 'react-native';
import { colors } from '../../theme/colors';

interface Props {
  endTime: Date;
  style?: StyleProp<ViewStyle>;
  onChange?(value: number): void;
  onFinish?(): void;
}

export default function SProgressTimeBar({ endTime, style, onChange, onFinish }: Props) {
  const progress = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    const now = new Date();
    const animation = Animated.timing(progress, {
      toValue: 1,
      duration: endTime.getTime() - now.getTime(),
      useNativeDriver: false,
    });

    animation.start(({ finished }) => {
      if (finished) onFinish?.();
    });

    if (onChange) {
      progress.addListener(({ value }) => {
        onChange?.(value);
      });
    }

    return () => {
      animation.stop();
      progress.removeAllListeners();
    };
  }, [endTime]);

  const width = progress.interpolate({
    inputRange: [0, 1],
    outputRange: ['0%', '100%'],
  });

  return (
    <View
      style={[
        { height: 5, backgroundColor: colors.gray, borderRadius: 10, overflow: 'hidden' },
        style,
      ]}
    >
      <Animated.View
        style={{
          width: width,
          backgroundColor: colors.primary,
          height: '100%',
        }}
      />
    </View>
  );

  return <div>SProgressTimeBar</div>;
}
