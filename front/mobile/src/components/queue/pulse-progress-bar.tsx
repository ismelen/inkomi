import React, { useEffect, useRef } from 'react';
import { Animated, Dimensions, StyleSheet, View } from 'react-native';
import { colors } from '../../theme/colors';

interface ProgressBarProps {
  progress: number; // valor entre 0 y 1
}

const { width: SCREEN_WIDTH } = Dimensions.get('window');

const clamp = (v: number) => Math.min(Math.max(v, 0), 1);

export default function PulseProgressBar({ progress }: ProgressBarProps) {
  const animatedWidth = useRef(new Animated.Value(0)).current;
  const sweepAnim = useRef(new Animated.Value(0)).current;
  const containerWidthRef = useRef(0);

  // Cuando el contenedor es medido: posicionamos la barra al valor actual
  // de forma instantánea (sin animación) para evitar la carrera cold-mount.
  const handleLayout = (e: any) => {
    const w = e.nativeEvent.layout.width;
    if (w <= 0 || w === containerWidthRef.current) return;
    containerWidthRef.current = w;
    animatedWidth.setValue(clamp(progress) * w);
  };

  // Cuando cambia el progress: animamos al nuevo valor en píxeles reales
  useEffect(() => {
    if (containerWidthRef.current === 0) return;
    Animated.timing(animatedWidth, {
      toValue: clamp(progress) * containerWidthRef.current,
      duration: 400,
      useNativeDriver: false,
    }).start();
  }, [progress]);

  // Loop del sweep (brillo)
  useEffect(() => {
    Animated.loop(
      Animated.timing(sweepAnim, {
        toValue: 1,
        duration: 1400,
        useNativeDriver: true,
      }),
    ).start();
  }, []);

  const translateX = sweepAnim.interpolate({
    inputRange: [0, 1],
    outputRange: [-150, SCREEN_WIDTH],
  });

  return (
    <View style={styles.container} onLayout={handleLayout}>
      <Animated.View style={[styles.progressBar, { width: animatedWidth }]}>
        <Animated.View style={[styles.sweepOverlay, { transform: [{ translateX }] }]} />
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
