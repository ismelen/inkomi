import { useEffect, useRef } from 'react';
import { Animated, StyleSheet, View } from 'react-native';
import { colors, hexToRgba } from '../../theme/colors';

function PulseBlock({ style }: { style?: any }) {
  const opacity = useRef(new Animated.Value(0.3)).current;

  useEffect(() => {
    const loop = Animated.loop(
      Animated.sequence([
        Animated.timing(opacity, {
          toValue: 1,
          duration: 700,
          useNativeDriver: true,
        }),
        Animated.timing(opacity, {
          toValue: 0.3,
          duration: 700,
          useNativeDriver: true,
        }),
      ])
    );
    loop.start();
    return () => loop.stop();
  }, [opacity]);

  return <Animated.View style={[{ opacity }, style]} />;
}

export default function SearchedBookCardSkeleton() {
  return (
    <View
      style={{
        borderRadius: 12,
        backgroundColor: colors.surface_container_low,
        padding: 10,
        marginTop: 10,
        flexDirection: 'row',
        gap: 10,
      }}
    >
      <PulseBlock
        style={{
          height: 100,
          borderRadius: 8,
          aspectRatio: 0.67,
          backgroundColor: hexToRgba(colors.on_surface_variant, 0.25),
        }}
      />

      <View style={{ flex: 1, overflow: 'hidden', justifyContent: 'space-between' }}>
        <View>
          <PulseBlock
            style={{
              height: 16,
              width: '85%',
              borderRadius: 4,
              backgroundColor: hexToRgba(colors.on_surface_variant, 0.25),
            }}
          />
          <PulseBlock
            style={{
              height: 16,
              width: '60%',
              borderRadius: 4,
              marginTop: 4,
              backgroundColor: hexToRgba(colors.on_surface_variant, 0.25),
            }}
          />
        </View>

        <View style={{ justifyContent: 'space-between', marginTop: 10 }}>
          <PulseBlock
            style={[
              styles.label,
              { width: 90, height: 16, backgroundColor: colors.tertiary_fixed_dim },
            ]}
          />

          <View style={{ flexDirection: 'row', marginTop: 8, gap: 8 }}>
            <PulseBlock
              style={[
                styles.label,
                { width: 40, height: 16, backgroundColor: colors.tertiary_fixed },
              ]}
            />
            <PulseBlock
              style={[
                styles.label,
                { width: 60, height: 16, backgroundColor: colors.secondary_fixed },
              ]}
            />
          </View>
        </View>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  label: {
    borderRadius: 8,
  },
});
