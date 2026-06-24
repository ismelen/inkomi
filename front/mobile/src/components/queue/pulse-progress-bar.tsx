import React, { useEffect, useRef } from 'react';
import { StyleSheet, View, Animated, Dimensions } from 'react-native';
import { colors } from '../../theme/colors';

interface ProgressBarProps {
  progress: number;
}

const { width: SCREEN_WIDTH } = Dimensions.get('window');

export default function PulseProgressBar({ progress }: ProgressBarProps) {
  const animatedProgress = useRef(new Animated.Value(0)).current;
  const sweepAnim = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    Animated.timing(animatedProgress, {
      toValue: Math.min(Math.max(progress, 0), 1),
      duration: 400,
      useNativeDriver: false,
    }).start();
  }, [progress]);

  useEffect(() => {
    Animated.loop(
      Animated.timing(sweepAnim, {
        toValue: 1,
        duration: 1000,
        useNativeDriver: true,
      })
    ).start();
  }, []);

  const width = animatedProgress.interpolate({
    inputRange: [0, 1],
    outputRange: ['0%', '100%'],
  });

  const translateX = sweepAnim.interpolate({
    inputRange: [0, 1],
    outputRange: [-150, SCREEN_WIDTH],
  });

  return (
    <View style={styles.container}>
      <Animated.View style={[styles.progressBar, { width }]}>
        <Animated.View
          style={[
            styles.sweepOverlay,
            {
              transform: [{ translateX }],
            },
          ]}
        />
      </Animated.View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    height: 4,
    width: '100%',
    backgroundColor: colors.primary_fixed,
    borderRadius: 6,
    overflow: 'hidden',
  },
  progressBar: {
    height: '100%',
    backgroundColor: colors.primary,
    borderRadius: 6,
    overflow: 'hidden',
    position: 'relative',
  },
  sweepOverlay: {
    position: 'absolute',
    top: 0,
    bottom: 0,
    width: 100,
    backgroundColor: 'rgba(255, 255, 255, 0.4)',
  },
});
