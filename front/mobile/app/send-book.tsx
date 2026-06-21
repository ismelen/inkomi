import React from 'react';
import { View } from 'react-native';
import SText from '../src/components/shared/SText';
import { Stack } from 'expo-router';
import { colors } from '../src/theme/colors';

export default function SendBookPage() {
  return (
    <>
      <Stack.Screen
        options={{
          headerShown: true,
          title: 'Send Book',
          headerTitleStyle: { fontFamily: 'semibold', fontSize: 20, color: colors.on_background },
          headerTitleAlign: 'center',
          headerStyle: {
            backgroundColor: colors.background,
          },
          headerTintColor: colors.primary,
        }}
      />
      <View>
        <SText>Send book</SText>
      </View>
    </>
  );
}
